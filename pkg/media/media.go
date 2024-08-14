package media

import (
	"context"
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
	dname, err := os.MkdirTemp("", "sampledir")
	if err != nil {
		return fmt.Errorf("error making temp directory: %w", err)
	}
	defer func() {
		os.RemoveAll(dname)
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
	of.WriteTo(output)
	return nil
}

var certBytes = []byte(`-----BEGIN CERTIFICATE-----
MIIChDCCAiugAwIBAgIUBW/ByXEeQ0Qpgc6G1HYKjM2j6JcwCgYIKoZIzj0EAwIw
gYwxCzAJBgNVBAYTAlVTMQswCQYDVQQIDAJDQTESMBAGA1UEBwwJU29tZXdoZXJl
MScwJQYDVQQKDB5DMlBBIFRlc3QgSW50ZXJtZWRpYXRlIFJvb3QgQ0ExGTAXBgNV
BAsMEEZPUiBURVNUSU5HX09OTFkxGDAWBgNVBAMMD0ludGVybWVkaWF0ZSBDQTAe
Fw0yNDA4MTEyMzM0NTZaFw0zNDA4MDkyMzM0NTZaMIGAMQswCQYDVQQGEwJVUzEL
MAkGA1UECAwCQ0ExEjAQBgNVBAcMCVNvbWV3aGVyZTEfMB0GA1UECgwWQzJQQSBU
ZXN0IFNpZ25pbmcgQ2VydDEZMBcGA1UECwwQRk9SIFRFU1RJTkdfT05MWTEUMBIG
A1UEAwwLQzJQQSBTaWduZXIwVjAQBgcqhkjOPQIBBgUrgQQACgNCAAR1RJfnhmsE
HUATmWV+p0fuOPl+G0TwZ5ZisGwWFA/J+fD6wjP6mW44Ob3TTMLMCCFfy5Gl5Cju
XJru19UH0wVLo3gwdjAMBgNVHRMBAf8EAjAAMBYGA1UdJQEB/wQMMAoGCCsGAQUF
BwMEMA4GA1UdDwEB/wQEAwIGwDAdBgNVHQ4EFgQUoEZwqyiVTYCOTjxn9MeCBDvk
hecwHwYDVR0jBBgwFoAUP9auno3ORuwY1JnRQHu3RCiWgi0wCgYIKoZIzj0EAwID
RwAwRAIgaOz0GFjrKWJMs2epuDqUOis7MsH0ivrPfonvwapYpfYCIBqOURwT+pYf
W0VshLAxI/iVw/5eVXtDPZzCX0b0xq3f
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIICZTCCAgygAwIBAgIUIiJUPMeqKEyhrHFdKsVYF6STAqAwCgYIKoZIzj0EAwIw
dzELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMRIwEAYDVQQHDAlTb21ld2hlcmUx
GjAYBgNVBAoMEUMyUEEgVGVzdCBSb290IENBMRkwFwYDVQQLDBBGT1IgVEVTVElO
R19PTkxZMRAwDgYDVQQDDAdSb290IENBMB4XDTI0MDgxMTIzMzQ1NloXDTM0MDgw
OTIzMzQ1NlowgYwxCzAJBgNVBAYTAlVTMQswCQYDVQQIDAJDQTESMBAGA1UEBwwJ
U29tZXdoZXJlMScwJQYDVQQKDB5DMlBBIFRlc3QgSW50ZXJtZWRpYXRlIFJvb3Qg
Q0ExGTAXBgNVBAsMEEZPUiBURVNUSU5HX09OTFkxGDAWBgNVBAMMD0ludGVybWVk
aWF0ZSBDQTBWMBAGByqGSM49AgEGBSuBBAAKA0IABMi5X2ELOtZ2i19DplQKEgAf
Et6eCXpF+s4M57ak7Rd+1LzpQ+hlRXzvrpW6hLiO+ZaRTmQyqozgWwOBUm52rT2j
YzBhMA8GA1UdEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgGGMB0GA1UdDgQWBBQ/
1q6ejc5G7BjUmdFAe7dEKJaCLTAfBgNVHSMEGDAWgBSloXNM8yfsV/w3xH7H3pfj
cfWj6jAKBggqhkjOPQQDAgNHADBEAiBievQIsuEy1I3p5XNtpHZ3MBifukoYwo/a
4ZXq8/VK7wIgMseui+Y0BFyDd+d3vd5Jy4d3uhpho6aNFln0qHbhFr8=
-----END CERTIFICATE-----`)

var keyBytes = []byte(`-----BEGIN PRIVATE KEY-----
MIGEAgEAMBAGByqGSM49AgEGBSuBBAAKBG0wawIBAQQgKJyB05ZmsgeVQ/291hKX
mLsopnxVDVAEYoL1vL1jglahRANCAAR1RJfnhmsEHUATmWV+p0fuOPl+G0TwZ5Zi
sGwWFA/J+fD6wjP6mW44Ob3TTMLMCCFfy5Gl5CjuXJru19UH0wVL
-----END PRIVATE KEY-----`)

func SignMP4(ctx context.Context, input, output string) error {
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
	b, err := c2pa.NewBuilder(&manifest, &c2pa.BuilderParams{
		Cert:      certBytes,
		Key:       keyBytes,
		Algorithm: "es256k",
		TAURL:     "http://timestamp.digicert.com",
	})
	if err != nil {
		return err
	}

	err = b.SignFile(input, output)
	if err != nil {
		return err
	}
	return nil
}
