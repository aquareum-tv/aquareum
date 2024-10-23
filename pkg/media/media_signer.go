package media

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"aquareum.tv/aquareum/pkg/aqio"
	"aquareum.tv/aquareum/pkg/aqtime"
	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/crypto/aqpub"
	"aquareum.tv/aquareum/pkg/crypto/signers"
	"aquareum.tv/aquareum/pkg/log"
	"git.aquareum.tv/aquareum-tv/c2pa-go/pkg/c2pa"
)

type MediaSigner struct {
	StreamerName string
	Signer       crypto.Signer
	Pub          aqpub.Pub
	Cert         []byte
	TAURL        string
}

func MakeMediaSigner(ctx context.Context, cli *config.CLI, streamer string, signer crypto.Signer) (*MediaSigner, error) {
	pub, err := aqpub.FromPublicKey(signer.Public().(*ecdsa.PublicKey))
	if err != nil {
		return nil, err
	}
	exists, err := cli.DataFileExists([]string{pub.String(), CERT_FILE})
	if err != nil {
		return nil, err
	}
	if !exists {
		cert, err := signers.GenerateES256KCert(signer)
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(cert)
		err = cli.DataFileWrite([]string{pub.String(), CERT_FILE}, r, false)
		if err != nil {
			return nil, err
		}
		log.Log(ctx, "wrote new media signing certificate", "file", filepath.Join(pub.String(), CERT_FILE))
	}
	buf := bytes.Buffer{}
	cli.DataFileRead([]string{pub.String(), CERT_FILE}, &buf)
	cert := buf.Bytes()
	return &MediaSigner{
		// cli:        cli,
		Signer:       signer,
		Cert:         cert,
		StreamerName: streamer,
		TAURL:        cli.TAURL,
		Pub:          pub,
	}, nil
}

func (ms *MediaSigner) SignMP4(ctx context.Context, input io.ReadSeeker, start int64) ([]byte, error) {
	end := time.Now().UnixMilli()
	mani := obj{
		"title": fmt.Sprintf("Livestream Segment at %s", aqtime.FromMillis(start)),
		"assertions": []obj{
			{
				"label": "c2pa.actions",
				"data": obj{
					"actions": []obj{
						{"action": "c2pa.created"},
						{"action": "c2pa.published"},
					},
				},
			},
			{
				"label": "stds.metadata",
				"data": obj{
					"@context": obj{
						"s": "http://schema.org/",
					},
					"@type": "s:VideoObject",
					"s:creator": []obj{
						{
							"@type":     "s:Person",
							"s:name":    ms.StreamerName,
							"s:address": ms.Pub.String(),
						},
					},
					"s:startTime": aqtime.FromMillis(start).String(),
					"s:endTime":   aqtime.FromMillis(end).String(),
				},
			},
		},
	}
	manifestBs, err := json.Marshal(mani)
	if err != nil {
		return nil, err
	}
	var manifest c2pa.ManifestDefinition
	err = json.Unmarshal(manifestBs, &manifest)
	if err != nil {
		return nil, err
	}
	alg, err := c2pa.GetSigningAlgorithm(string(c2pa.ES256K))
	if err != nil {
		return nil, err
	}
	b, err := c2pa.NewBuilder(&manifest, &c2pa.BuilderParams{
		Cert:      ms.Cert,
		Signer:    ms.Signer,
		Algorithm: alg,
		TAURL:     ms.TAURL,
	})
	if err != nil {
		return nil, err
	}

	output := &aqio.ReadWriteSeeker{}
	err = b.Sign(input, output, "video/mp4")
	if err != nil {
		return nil, err
	}
	bs, err := output.Bytes()
	if err != nil {
		return nil, err
	}
	return bs, nil
}
