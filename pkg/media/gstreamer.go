package media

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
	"golang.org/x/sync/errgroup"
)

func NormalizeAudio(ctx context.Context, input io.Reader, output io.Writer) error {
	// Initialize GStreamer with the arguments passed to the program. Gstreamer
	// and the bindings will automatically pop off any handled arguments leaving
	// nothing but a pipeline string (unless other invalid args are present).
	gst.Init(nil)
	ir, iw, err := os.Pipe()
	if err != nil {
		return err
	}
	or, ow, err := os.Pipe()
	if err != nil {
		return err
	}
	// Create a main loop. This is only required when utilizing signals via the bindings.
	// In this example, the AddWatch on the pipeline bus requires iterating on the main loop.
	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

	// Build a pipeline string from the cli arguments
	pipelineSlice := []string{
		fmt.Sprintf("fdsrc fd=%d", ir.Fd()),
		"matroskademux",
		"opusdec use-inband-fec=true",
		"audioresample",
		"fdkaacenc",
		"matroskamux",
		fmt.Sprintf("fdsink fd=%d", ow.Fd()),
	}
	pipelineString := strings.Join(pipelineSlice, " ! ")

	/// Let GStreamer create a pipeline from the parsed launch syntax on the cli.
	pipeline, err := gst.NewPipelineFromString(pipelineString)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	// Add a message handler to the pipeline bus, printing interesting information to the console.
	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS: // When end-of-stream is received flush the pipeling and stop the main loop
			pipeline.BlockSetState(gst.StateNull)
			mainLoop.Quit()
		case gst.MessageError: // Error messages are always fatal
			err := msg.ParseError()
			fmt.Println("ERROR:", err.Error())
			if debug := err.DebugString(); debug != "" {
				fmt.Println("DEBUG:", debug)
			}
			mainLoop.Quit()
		default:
			// All messages implement a Stringer. However, this is
			// typically an expensive thing to do and should be avoided.
			fmt.Println(msg)
		}
		return true
	})

	// Start the pipeline
	pipeline.SetState(gst.StatePlaying)
	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		_, err := io.Copy(iw, input)
		fmt.Println("input copy complete")
		iw.Close()
		return err
	})

	g.Go(func() error {
		mainLoop.Run()
		fmt.Println("main loop complete")
		ow.Close()
		return nil
	})

	g.Go(func() error {
		_, err := io.Copy(output, or)
		fmt.Println("output copy complete")
		return err
	})

	return g.Wait()
}
