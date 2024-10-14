package media

import (
	"bytes"
	"context"
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

func SelfTest(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	f, err := test.Files.Open("fixtures/sample-segment.mp4")
	if err != nil {
		return err
	}

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
		fmt.Sprintf("fdsrc fd=%d ! fdsink fd=%d", ir.Fd(), ow.Fd()),
	}

	pipeline, err := gst.NewPipelineFromString(strings.Join(pipelineSlice, "\n"))
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		pipeline.BlockSetState(gst.StateNull)
		mainLoop.Quit()
	}()

	// Start the pipeline
	pipeline.SetState(gst.StatePlaying)

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		_, err := io.Copy(iw, f)
		iw.Close()
		pipeline.BlockSetState(gst.StateNull)
		mainLoop.Quit()
		return err
	})

	g.Go(func() error {
		mainLoop.Run()
		ow.Close()
		return nil
	})

	var output bytes.Buffer
	g.Go(func() error {
		_, err := io.Copy(&output, or)
		return err
	})

	err = g.Wait()
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

func (mm *MediaManager) TestSource(ctx context.Context) error {
	or, ow, odone, err := SafePipe()
	if err != nil {
		return err
	}
	defer odone()

	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

	pipelineSlice := []string{
		fmt.Sprintf("matroskamux name=mux ! fdsink fd=%d", ow.Fd()),
		`videotestsrc is-live=true ! video/x-raw,width=1280,height=720 ! x264enc speed-preset=ultrafast key-int-max=30 ! mux.`,
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
		mainLoop.Run()
		log.Log(ctx, "main loop complete")
		return nil
	})

	g.Go(func() error {
		prefix := fmt.Sprintf("%s/segment/%s", mm.cli.OwnInternalURL(), "foo")
		err = SegmentToHTTP(ctx, or, prefix)
		log.Log(ctx, "output copy complete", "error", err)
		return err
	})

	return g.Wait()
}
