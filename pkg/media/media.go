package media

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/crypto/signers/eip712"
	"aquareum.tv/aquareum/pkg/log"
	"github.com/google/uuid"
	"github.com/livepeer/lpms/ffmpeg"
	"golang.org/x/sync/errgroup"

	"git.aquareum.tv/aquareum-tv/c2pa-go/pkg/c2pa"
)

const CERT_FILE = "cert.pem"
const SEGMENTS_DIR = "segments"

type MediaManager struct {
	cli        *config.CLI
	signer     crypto.Signer
	cert       []byte
	user       string
	mp4subs    map[string][]chan string
	mp4subsmut sync.Mutex
	mkvsubs    map[string]io.Writer
	mkvsubsmut sync.Mutex
}

func MakeMediaManager(ctx context.Context, cli *config.CLI, signer *eip712.EIP712Signer) (*MediaManager, error) {
	exists, err := cli.DataFileExists([]string{CERT_FILE})
	if err != nil {
		return nil, err
	}
	if !exists {
		cert, err := signer.GenerateCert()
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(cert)
		err = cli.DataFileWrite([]string{CERT_FILE}, r, false)
		if err != nil {
			return nil, err
		}
		log.Log(ctx, "wrote new media signing certificate", "file", CERT_FILE)
	}
	buf := bytes.Buffer{}
	cli.DataFileRead([]string{CERT_FILE}, &buf)
	cert := buf.Bytes()
	return &MediaManager{
		cli:     cli,
		signer:  signer,
		cert:    cert,
		user:    signer.Hex(),
		mp4subs: map[string][]chan string{},
		mkvsubs: map[string]io.Writer{},
	}, nil
}

// accept an incoming mkv segment, mux to mp4, and sign it
func (mm *MediaManager) SignSegment(ctx context.Context, input io.Reader, ms int64) error {
	segmentFile := fmt.Sprintf("%d.mp4", ms)
	buf := bytes.Buffer{}
	err := MuxToMP4(ctx, input, &buf)
	if err != nil {
		return fmt.Errorf("error muxing to mp4: %w", err)
	}
	reader := bytes.NewReader(buf.Bytes())
	fd, err := mm.cli.DataFileCreate([]string{SEGMENTS_DIR, mm.user, segmentFile}, false)
	if err != nil {
		return err
	}
	defer fd.Close()
	err = mm.SignMP4(ctx, reader, fd, ms)
	if err != nil {
		return fmt.Errorf("error signing mp4: %w", err)
	}
	go mm.PublishSegment(ctx, mm.user, segmentFile)
	return nil
}

// subscribe to the latest segments from a given user for livestreaming purposes
func (mm *MediaManager) SubscribeSegment(ctx context.Context, user string) chan string {
	mm.mp4subsmut.Lock()
	defer mm.mp4subsmut.Unlock()
	_, ok := mm.mp4subs[user]
	if !ok {
		mm.mp4subs[user] = []chan string{}
	}
	c := make(chan string)
	mm.mp4subs[user] = append(mm.mp4subs[user], c)
	return c
}

// subscribe to the latest segments from a given user for livestreaming purposes
func (mm *MediaManager) PublishSegment(ctx context.Context, user, file string) {
	mm.mp4subsmut.Lock()
	defer mm.mp4subsmut.Unlock()
	for _, sub := range mm.mp4subs[user] {
		sub <- file
	}
	mm.mp4subs[user] = []chan string{}
}

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
		// log.Log(ctx, "cleaning up")
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
				Opts: map[string]string{"movflags": "+faststart"},
			},
		},
	}
	iname := fmt.Sprintf("pipe:%d", ir.Fd())
	in := &ffmpeg.TranscodeOptionsIn{Fname: iname, Transmuxing: true}
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		_, err := io.Copy(iw, input)
		// log.Log(ctx, "input copy done", "error", err)
		iw.Close()
		return err
	})
	g.Go(func() error {
		_, err = tc.Transcode(in, out)
		// log.Log(ctx, "transcode done", "error", err)
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
	_, err = io.Copy(output, of)
	if err != nil {
		return err
	}
	of.Close()
	status, info, err := ffmpeg.GetCodecInfo(oname)
	if err != nil {
		return fmt.Errorf("error in GetCodecInfo: %w", err)
	}
	fmt.Printf("%v %v\n", status, info.DurSecs)
	// log.Log(ctx, "transmuxing complete", "out-file", oname, "wrote", written)
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
		// log.Log(ctx, "input copy done", "error", err)
		iw.Close()
		return err
	})
	g.Go(func() error {
		_, err = tc.Transcode(in, out)
		// log.Log(ctx, "transcode done", "error", err)
		tc.StopTranscoder()
		ir.Close()
		return err
	})
	return g.Wait()
}

func (mm *MediaManager) StreamToMKV(ctx context.Context, user string, w io.Writer) error {
	tc := ffmpeg.NewTranscoder()
	defer tc.StopTranscoder()
	uu, err := uuid.NewV7()
	if err != nil {
		return err
	}
	mm.mkvsubsmut.Lock()
	mm.mkvsubs[uu.String()] = w
	mm.mkvsubsmut.Unlock()
	iname := fmt.Sprintf("%s/playback/%s/concat", mm.cli.OwnInternalURL(), user)
	in := &ffmpeg.TranscodeOptionsIn{
		Fname:       iname,
		Transmuxing: true,
		Profile:     ffmpeg.VideoProfile{},
		Loop:        -1,
		Demuxer: ffmpeg.ComponentOptions{
			Name: "concat",
			Opts: map[string]string{
				"safe":               "0",
				"protocol_whitelist": "file,http,https,tcp,tls",
			},
		},
	}
	out := []ffmpeg.TranscodeOptions{
		{
			Oname: fmt.Sprintf("%s/playback/%s/%s/stream.mkv", mm.cli.OwnInternalURL(), user, uu.String()),
			VideoEncoder: ffmpeg.ComponentOptions{
				Name: "copy",
			},
			AudioEncoder: ffmpeg.ComponentOptions{
				Name: "copy",
			},
			Profile: ffmpeg.VideoProfile{Format: ffmpeg.FormatNone},
			Muxer: ffmpeg.ComponentOptions{
				Name: "matroska",
			},
		},
	}
	_, err = tc.Transcode(in, out)
	return err
}

func (mm *MediaManager) HandleMKVStream(ctx context.Context, user, uu string, r io.Reader) error {
	mm.mkvsubsmut.Lock()
	w, ok := mm.mkvsubs[uu]
	mm.mkvsubsmut.Unlock()
	if !ok {
		return fmt.Errorf("uuid not found: %s", uu)
	}
	_, err := io.Copy(w, r)
	return err
}

func (mm *MediaManager) SignMP4(ctx context.Context, input io.ReadSeeker, output io.ReadWriteSeeker, now int64) error {
	manifestBs := []byte(fmt.Sprintf(`
		{
			"title": "Livestream Segment at %s",
			"assertions": [
				{
					"label": "c2pa.actions",
					"data": {"actions": [
						{ "action": "c2pa.created" },
						{ "action": "c2pa.published" }
					]}
				}
			]
		}
	`, time.UnixMilli(now).UTC().Format("2006-01-02T15:04:05.999Z")))
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
		Cert:      mm.cert,
		Signer:    mm.signer,
		Algorithm: alg,
		TAURL:     mm.cli.TAURL,
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
