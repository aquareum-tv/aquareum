package media

import (
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"aquareum.tv/aquareum/pkg/log"
	"github.com/livepeer/lpms/ffmpeg"
	"golang.org/x/sync/errgroup"

	"git.aquareum.tv/aquareum-tv/c2pa-go/pkg/c2pa"
)

func MuxToMP4(ctx context.Context, input io.Reader, output io.Writer) error {
	tc := ffmpeg.NewTranscoder()
	ir, iw, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("error opening pipe: %w", err)
	}
	dname, err := os.MkdirTemp("", "aquareum-muxing")
	if err != nil {
		return fmt.Errorf("error making temp directory: %w", err)
	}
	defer func() {
		log.Log(ctx, "cleaning up")
		tc.StopTranscoder()
	}()
	oname := filepath.Join(dname, "output.mp4")
	out := []ffmpeg.TranscodeOptions{
		{
			Oname: oname,
			VideoEncoder: ffmpeg.ComponentOptions{
				Name: "copy",
			},
			AudioEncoder: ffmpeg.ComponentOptions{
				Name: "copy",
			},
			Profile: ffmpeg.VideoProfile{Format: ffmpeg.FormatNone},
			Muxer: ffmpeg.ComponentOptions{
				Name: "mp4",
				// main option is 'frag_keyframe' which tells ffmpeg to create fragmented MP4 (which we need to be able to stream generatd file)
				// other options is not mandatory but they will slightly improve generated MP4 file
				Opts: map[string]string{"movflags": "+faststart"},
				// Opts: map[string]string{"movflags": "frag_keyframe+negative_cts_offsets+omit_tfhd_offset+disable_chpl+default_base_moof"},
			},
		},
	}
	iname := fmt.Sprintf("pipe:%d", ir.Fd())
	in := &ffmpeg.TranscodeOptionsIn{Fname: iname, Transmuxing: true}
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		_, err := io.Copy(iw, input)
		log.Log(ctx, "input copy done", "error", err)
		iw.Close()
		return err
	})
	g.Go(func() error {
		_, err = tc.Transcode(in, out)
		log.Log(ctx, "transcode done", "error", err)
		tc.StopTranscoder()
		ir.Close()
		return err
	})
	err = g.Wait()
	if err != nil {
		return err
	}
	of, err := os.Open(oname)
	if err != nil {
		return err
	}
	defer of.Close()
	written, err := io.Copy(output, of)
	if err != nil {
		return err
	}
	log.Log(ctx, "transmuxing complete", "out-file", oname, "wrote", written)
	return nil
}

func SegmentToHTTP(ctx context.Context, input io.Reader, prefix string) error {
	tc := ffmpeg.NewTranscoder()
	defer tc.StopTranscoder()
	ir, iw, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("error opening pipe: %w", err)
	}
	out := []ffmpeg.TranscodeOptions{
		{
			Oname: fmt.Sprintf("%s/%%d.mkv", prefix),
			VideoEncoder: ffmpeg.ComponentOptions{
				Name: "copy",
			},
			AudioEncoder: ffmpeg.ComponentOptions{
				Name: "copy",
			},
			Profile: ffmpeg.VideoProfile{Format: ffmpeg.FormatNone},
			Muxer: ffmpeg.ComponentOptions{
				Name: "stream_segment",
				Opts: map[string]string{
					"segment_time": "0.1",
				},
			},
		},
	}
	iname := fmt.Sprintf("pipe:%d", ir.Fd())
	in := &ffmpeg.TranscodeOptionsIn{Fname: iname, Transmuxing: true}
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		_, err := io.Copy(iw, input)
		log.Log(ctx, "input copy done", "error", err)
		iw.Close()
		return err
	})
	g.Go(func() error {
		_, err = tc.Transcode(in, out)
		log.Log(ctx, "transcode done", "error", err)
		tc.StopTranscoder()
		ir.Close()
		return err
	})
	return g.Wait()
}

func SignMP4(ctx context.Context, signer crypto.Signer, certBytes []byte, input io.ReadSeeker, output io.ReadWriteSeeker) error {
	manifestBs := []byte(`
		{
			"title": "Image File",
			"assertions": [
				{
					"label": "c2pa.actions",
					"data": { "actions": [{ "action": "c2pa.published" }] }
				}
			]
		}
	`)
	var manifest c2pa.ManifestDefinition
	err := json.Unmarshal(manifestBs, &manifest)
	if err != nil {
		return err
	}
	alg, err := c2pa.GetSigningAlgorithm(string(c2pa.ES256K))
	if err != nil {
		return err
	}
	b, err := c2pa.NewBuilder(&manifest, &c2pa.BuilderParams{
		Cert:      certBytes,
		Signer:    signer,
		Algorithm: alg,
		TAURL:     "http://timestamp.digicert.com",
	})
	if err != nil {
		return err
	}

	err = b.Sign(input, output, "video/mp4")
	if err != nil {
		return err
	}
	return nil
}
