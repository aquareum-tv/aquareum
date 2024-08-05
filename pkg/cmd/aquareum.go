package cmd

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"aquareum.tv/aquareum/pkg/crypto/signers/eip712"
	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/notifications"
	"aquareum.tv/aquareum/pkg/proc"
	v0 "aquareum.tv/aquareum/pkg/schema/v0"

	"aquareum.tv/aquareum/pkg/api"
	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/model"
	"github.com/adrg/xdg"
	"github.com/peterbourgon/ff/v3"
)

// parse the CLI and fire up an aquareum node!
func Start(build *config.BuildFlags) error {
	if os.Args[1] == "slurp-file" {
		fs := flag.NewFlagSet("aquareum-slurp-file", flag.ExitOnError)
		inurl := fs.String("url", "", "Base URL to send slurped files to")
		fname := fs.String("file", "", "Name of this file we're uploading")
		ff.Parse(
			fs, os.Args[2:],
			ff.WithEnvVarPrefix("AQ"),
		)
		*fname = strings.TrimPrefix(*fname, config.AQUAREUM_SCHEME_PREFIX)
		fmt.Printf("file slurpin args=%s\n", strings.Join(os.Args, ", "))

		fullURL := fmt.Sprintf("%s/segment/%s", *inurl, *fname)

		reader := bufio.NewReader(os.Stdin)
		req, err := http.NewRequest("POST", fullURL, reader)
		if err != nil {
			panic(err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		fmt.Printf("http %s\n", resp.Status)
		os.Exit(0)
	}
	err := normalizeXDG()
	if err != nil {
		return err
	}

	tlsCertFile, err := xdg.ConfigFile("aquareum/tls/tls.crt")
	if err != nil {
		return err
	}
	tlsKeyFile, err := xdg.ConfigFile("aquareum/tls/tls.key")
	if err != nil {
		return err
	}
	dbFile, err := xdg.DataFile("aquareum/db.sqlite")
	if err != nil {
		return err
	}
	dbFile = fmt.Sprintf("sqlite://%s", dbFile)

	fs := flag.NewFlagSet("aquareum", flag.ExitOnError)
	cli := config.CLI{Build: build}
	fs.StringVar(&cli.HttpAddr, "http-addr", ":8080", "Public HTTP address")
	fs.StringVar(&cli.HttpInternalAddr, "http-internal-addr", "127.0.0.1:9090", "Private, admin-only HTTP address")
	fs.StringVar(&cli.HttpsAddr, "https-addr", ":8443", "Public HTTPS address")
	fs.BoolVar(&cli.Insecure, "insecure", false, "Run without HTTPS. not recomended, as WebRTC support requires HTTPS")
	fs.StringVar(&cli.TLSCertPath, "tls-cert", tlsCertFile, "Path to TLS certificate")
	fs.StringVar(&cli.TLSKeyPath, "tls-key", tlsKeyFile, "Path to TLS key")
	fs.StringVar(&cli.SigningKeyPath, "signing-key", "", "Path to signing key for pushing OTA updates to the app")
	fs.StringVar(&cli.DBPath, "db-path", dbFile, "path to sqlite database file")
	fs.StringVar(&cli.AdminAccount, "admin-account", "", "ethereum account that administrates this aquareum node")
	fs.StringVar(&cli.FirebaseServiceAccount, "firebase-service-account", "", "JSON string of a firebase service account key")
	fs.IntVar(&cli.MistAdminPort, "mist-admin-port", 14242, "MistServer admin port (internal use only)")
	fs.IntVar(&cli.MistRTMPPort, "mist-rtmp-port", 11935, "MistServer RTMP port (internal use only)")
	fs.IntVar(&cli.MistHTTPPort, "mist-http-port", 18080, "MistServer HTTP port (internal use only)")
	fs.StringVar(&cli.GitLabURL, "gitlab-url", "https://git.aquareum.tv/api/v4/projects/1", "gitlab url for generating download links")
	version := fs.Bool("version", false, "print version and exit")

	ff.Parse(
		fs, os.Args[1:],
		ff.WithEnvVarPrefix("AQ"),
	)

	log.Log(context.Background(),
		"aquareum",
		"version", build.Version,
		"buildTime", build.BuildTimeStr(),
		"uuid", build.UUID,
		"runtime.GOOS", runtime.GOOS,
		"runtime.GOARCH", runtime.GOARCH)
	if *version {
		return nil
	}

	schema, err := v0.MakeV0Schema()
	if err != nil {
		return err
	}
	signer, err := eip712.MakeEIP712Signer(context.Background(), &eip712.EIP712SignerOptions{
		Schema: schema,
	})
	if err != nil {
		return err
	}
	mod, err := model.MakeDB(cli.DBPath)
	if err != nil {
		return err
	}
	var noter notifications.FirebaseNotifier
	if cli.FirebaseServiceAccount != "" {
		noter, err = notifications.MakeFirebaseNotifier(context.Background(), cli.FirebaseServiceAccount)
		if err != nil {
			return err
		}
	}
	a, err := api.MakeAquareumAPI(&cli, mod, signer, noter)
	if err != nil {
		return err
	}

	group, ctx := TimeoutGroupWithContext(context.Background())
	ctx = log.WithLogValues(ctx, "version", build.Version)

	group.Go(func() error {
		return handleSignals(ctx)
	})

	if !cli.Insecure {
		group.Go(func() error {
			return a.ServeHTTPS(ctx)
		})
		group.Go(func() error {
			return a.ServeHTTPRedirect(ctx)
		})
	} else {
		group.Go(func() error {
			return a.ServeHTTP(ctx)
		})
	}

	group.Go(func() error {
		return a.ServeInternalHTTP(ctx)
	})

	group.Go(func() error {
		return proc.RunMistServer(ctx, &cli)
	})

	return group.Wait()
}

// xdg sometimes gets confused in systemd, give it a default
func normalizeXDG() error {
	if xdg.Home == "/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		os.Setenv("HOME", home)
		xdg.Reload()
	}
	if xdg.Home == "/" {
		return fmt.Errorf("couldn't find home directory")
	}
	return nil
}

func handleSignals(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case s := <-c:
			log.Log(ctx, "caught signal, attempting clean shutdown", "signal", s)
			return fmt.Errorf("caught signal=%v", s)
		case <-ctx.Done():
			return nil
		}
	}
}
