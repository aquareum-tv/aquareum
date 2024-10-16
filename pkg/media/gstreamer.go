package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/test"
	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"github.com/skip2/go-qrcode"
	"golang.org/x/sync/errgroup"
)

const HLS_PLAYLIST = "stream.m3u8"

// Pipe with a mechanism to keep the FDs not garbage collected
func SafePipe() (*os.File, *os.File, func(), error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, nil, err
	}
	return r, w, func() {
		runtime.KeepAlive(r.Fd())
		runtime.KeepAlive(w.Fd())
	}, nil
}

func AddOpusToMKV(ctx context.Context, input io.Reader, output io.Writer) error {
	ir, iw, idone, err := SafePipe()
	if err != nil {
		return err
	}
	defer idone()
	or, ow, odone, err := SafePipe()
	if err != nil {
		return err
	}
	defer odone()

	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

	pipelineSlice := []string{
		fmt.Sprintf("fdsrc name=livestream fd=%d ! matroskademux name=demux", ir.Fd()),
		fmt.Sprintf("matroskamux name=mux ! fdsink fd=%d", ow.Fd()),
		"demux.audio_0 ! queue ! tee name=asplit",
		"demux.video_0 ! queue ! mux.video_0",
		"asplit. ! queue ! fdkaacdec ! audioresample ! opusenc inband-fec=true perfect-timestamp=true bitrate=128000 ! mux.audio_1",
		"asplit. ! queue ! mux.audio_0",
	}

	pipeline, err := gst.NewPipelineFromString(strings.Join(pipelineSlice, "\n"))
	if err != nil {
		return err
	}
	// demux, err := pipeline.GetElementByName("demux")
	// if err != nil {
	// 	return err
	// }
	// // Get the audiotestsrc's src-pad.
	// demuxPad := demux.GetStaticPad("sink")
	// if demuxPad == nil {
	// 	return fmt.Errorf("src pad on src element was nil")
	// }

	// // Add a probe handler on the audiotestsrc's src-pad.
	// // This handler gets called for every buffer that passes the pad we probe.
	// demuxPad.AddProbe(gst.PadProbeTypeAllBoth, func(self *gst.Pad, info *gst.PadProbeInfo) gst.PadProbeReturn {
	// 	// fmt.Printf("%v\n", info)
	// 	return gst.PadProbeOK
	// })

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		pipeline.BlockSetState(gst.StateNull)
		mainLoop.Quit()
	}()

	// Add a message handler to the pipeline bus, printing interesting information to the console.
	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {

		case gst.MessageEOS: // When end-of-stream is received flush the pipeling and stop the main loop
			cancel()
		case gst.MessageError: // Error messages are always fatal
			err := msg.ParseError()
			log.Log(ctx, "gstreamer error", "error", err.Error())
			if debug := err.DebugString(); debug != "" {
				log.Log(ctx, "gstreamer debug", "message", debug)
			}
			cancel()
		default:
			log.Log(ctx, msg.String())
		}
		return true
	})

	// Start the pipeline
	pipeline.SetState(gst.StatePlaying)

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		_, err := io.Copy(iw, input)
		log.Log(ctx, "input copy complete", "error", err)
		iw.Close()
		return err
	})

	g.Go(func() error {
		mainLoop.Run()
		log.Log(ctx, "main loop complete")
		ow.Close()
		return nil
	})

	g.Go(func() error {
		runtime.GC()
		_, err := io.Copy(output, or)
		log.Log(ctx, "output copy complete", "error", err)
		return err
	})

	return g.Wait()
}

// basic test to make sure gstreamer functionality is working
func SelfTest(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	f, err := test.Files.Open("fixtures/sample-segment.mp4")
	if err != nil {
		return err
	}
	defer f.Close()
	bs, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	pipeline, err := gst.NewPipelineFromString("appsrc name=src ! appsink name=sink")
	if err != nil {
		return err
	}

	srcele, err := pipeline.GetElementByName("src")
	if err != nil {
		return err
	}
	if srcele == nil {
		return fmt.Errorf("srcele not found")
	}
	src := app.SrcFromElement(srcele)
	src.SetCallbacks(&app.SourceCallbacks{
		NeedDataFunc: func(self *app.Source, _ uint) {
			buffer := gst.NewBufferWithSize(int64(len(bs)))
			buffer.Map(gst.MapWrite).WriteData(bs)
			self.PushBuffer(buffer)
			pipeline.SendEvent(gst.NewEOSEvent())
		},
	})

	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

	output := &bytes.Buffer{}
	sinkele, err := pipeline.GetElementByName("sink")
	if err != nil {
		return err
	}
	if sinkele == nil {
		return fmt.Errorf("sinkele not found")
	}
	appsink := app.SinkFromElement(sinkele)
	appsink.SetCallbacks(&app.SinkCallbacks{
		NewSampleFunc: func(sink *app.Sink) gst.FlowReturn {
			sample := sink.PullSample()
			if sample == nil {
				return gst.FlowOK
			}
			// defer sample.Unref()

			// Retrieve the buffer from the sample.
			buffer := sample.GetBuffer()

			_, err := io.Copy(output, buffer.Reader())

			if err != nil {
				panic(err)
			}

			return gst.FlowOK
		},
		EOSFunc: func(sink *app.Sink) {
			cancel()
		},
	})

	// Start the pipeline
	pipeline.SetState(gst.StatePlaying)

	go func() {
		<-ctx.Done()
		mainLoop.Quit()
	}()

	mainLoop.Run()

	if err != nil {
		return err
	}
	if len(output.Bytes()) < 1 {
		return fmt.Errorf("got a zero-byte buffer from SelfTest")
	}
	return nil
}

func ToHLS(ctx context.Context, input io.Reader, dir string) error {
	ir, iw, idone, err := SafePipe()
	if err != nil {
		return err
	}
	defer idone()

	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

	seg := filepath.Join(dir, "segment%05d.ts")
	playlist := filepath.Join(dir, HLS_PLAYLIST)
	pipelineSlice := []string{
		fmt.Sprintf("fdsrc name=livestream fd=%d ! matroskademux name=demux", ir.Fd()),
		fmt.Sprintf("hlssink2 name=mux location=%s target-duration=1 playlist-location=%s", seg, playlist),
		"demux.video_0 ! queue ! h264parse ! mux.video",
		"demux.audio_0 ! queue ! aacparse ! mux.audio",
	}

	pipeline, err := gst.NewPipelineFromString(strings.Join(pipelineSlice, "\n"))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		pipeline.BlockSetState(gst.StateNull)
		mainLoop.Quit()
	}()

	// Add a message handler to the pipeline bus, printing interesting information to the console.
	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {

		case gst.MessageEOS: // When end-of-stream is received flush the pipeling and stop the main loop
			cancel()
		case gst.MessageError: // Error messages are always fatal
			err := msg.ParseError()
			log.Log(ctx, "gstreamer error", "error", err.Error())
			if debug := err.DebugString(); debug != "" {
				log.Log(ctx, "gstreamer debug", "message", debug)
			}
			cancel()
		default:
			log.Log(ctx, msg.String())
		}
		return true
	})

	// Start the pipeline
	pipeline.SetState(gst.StatePlaying)

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		_, err := io.Copy(iw, input)
		log.Log(ctx, "input copy complete", "error", err)
		iw.Close()
		return err
	})

	g.Go(func() error {
		mainLoop.Run()
		log.Log(ctx, "main loop complete")
		return nil
	})

	return g.Wait()
}

const TESTSRC_WIDTH = 1280
const TESTSRC_HEIGHT = 720
const QR_SIZE = 256

type QRData struct {
	Now int64 `json:"now"`
}

func (mm *MediaManager) TestSource(ctx context.Context) error {
	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

	pipelineSlice := []string{
		"h264parse name=mux ! splitmuxsink name=splitter async-finalize=true sink-factory=appsink muxer-factory=matroskamux max-size-bytes=1",
		"compositor name=comp ! videoconvert ! x264enc speed-preset=ultrafast key-int-max=30 ! mux.",
		fmt.Sprintf(`videotestsrc is-live=true ! video/x-raw,format=AYUV,framerate=30/1,width=%d,height=%d ! comp.`, TESTSRC_WIDTH, TESTSRC_HEIGHT),
		fmt.Sprintf("videobox border-alpha=0 top=-%d left=-%d name=box ! comp.", (TESTSRC_HEIGHT/2)-(QR_SIZE/2), (TESTSRC_WIDTH/2)-(QR_SIZE/2)),
		"appsrc name=pngsrc ! pngdec ! videoconvert ! videorate ! video/x-raw,format=AYUV,framerate=1/1 ! box.",
	}

	pipeline, err := gst.NewPipelineFromString(strings.Join(pipelineSlice, "\n"))
	if err != nil {
		return err
	}

	ele, err := pipeline.GetElementByName("splitter")
	if err != nil {
		return err
	}
	if ele == nil {
		return fmt.Errorf("splitter not found")
	}

	ele.Connect("sink-added", func(split, sinkEle *gst.Element) {
		buf := &bytes.Buffer{}
		appsink := app.SinkFromElement(sinkEle)
		if appsink == nil {
			panic("appsink should not be nil")
		}
		appsink.SetCallbacks(&app.SinkCallbacks{
			NewSampleFunc: func(sink *app.Sink) gst.FlowReturn {
				sample := sink.PullSample()
				if sample == nil {
					return gst.FlowOK
				}
				// defer sample.Unref()

				// Retrieve the buffer from the sample.
				buffer := sample.GetBuffer()

				_, err := io.Copy(buf, buffer.Reader())

				if err != nil {
					panic(err)
				}

				return gst.FlowOK
			},
			EOSFunc: func(sink *app.Sink) {
				mm.SignSegment(ctx, buf, time.Now().UnixMilli())
			},
		})
	})

	pngele, err := pipeline.GetElementByName("pngsrc")
	if err != nil {
		return err
	}
	if pngele == nil {
		return fmt.Errorf("pngsrc not found")
	}
	src := app.SrcFromElement(pngele)
	src.SetCallbacks(&app.SourceCallbacks{
		NeedDataFunc: func(self *app.Source, _ uint) {
			now := time.Now().UnixMilli()
			data := QRData{Now: now}
			bs, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}
			png, err := qrcode.Encode(string(bs), qrcode.Medium, 256)
			if err != nil {
				panic(err)
			}
			buffer := gst.NewBufferWithSize(int64(len(png)))
			buffer.Map(gst.MapWrite).WriteData(png)
			self.PushBuffer(buffer)
		},
	})
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		pipeline.BlockSetState(gst.StateNull)
		mainLoop.Quit()
	}()

	// Add a message handler to the pipeline bus, printing interesting information to the console.
	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {

		case gst.MessageEOS: // When end-of-stream is received flush the pipeling and stop the main loop
			cancel()
		case gst.MessageError: // Error messages are always fatal
			err := msg.ParseError()
			log.Log(ctx, "gstreamer error", "error", err.Error())
			if debug := err.DebugString(); debug != "" {
				log.Log(ctx, "gstreamer debug", "message", debug)
			}
			cancel()
			// default:
			// 	log.Log(ctx, msg.String())
		}
		return true
	})

	// Start the pipeline
	pipeline.SetState(gst.StatePlaying)

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		mainLoop.Run()
		log.Log(ctx, "main loop complete")
		return nil
	})

	return g.Wait()
}
