package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"

	"aquareum.tv/aquareum/pkg/crypto/signers/eip712"
	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/media"
	"aquareum.tv/aquareum/pkg/notifications"
	v0 "aquareum.tv/aquareum/pkg/schema/v0"

	"aquareum.tv/aquareum/pkg/api"
	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/model"
	_ "github.com/go-gst/go-glib/glib"
	_ "github.com/go-gst/go-gst/gst"
)

// Additional jobs that can be injected by platforms
type jobFunc func(ctx context.Context, cli *config.CLI) error

// parse the CLI and fire up an aquareum node!
func start(build *config.BuildFlags, platformJobs []jobFunc) error {
	if len(os.Args) > 1 && os.Args[1] == "stream" {
		if len(os.Args) != 3 {
			fmt.Println("usage: aquareum stream [user]")
			os.Exit(1)
		}
		return Stream(os.Args[2])
	}

	defaultDataDir, err := config.DefaultDataDir()
	if err != nil {
		return err
	}

	fs := flag.NewFlagSet("aquareum", flag.ExitOnError)
	cli := config.CLI{Build: build}
	fs.StringVar(&cli.DataDir, "data-dir", defaultDataDir, "directory for keeping all aquareum data")
	fs.StringVar(&cli.HttpAddr, "http-addr", ":8080", "Public HTTP address")
	fs.StringVar(&cli.HttpInternalAddr, "http-internal-addr", "127.0.0.1:9090", "Private, admin-only HTTP address")
	fs.StringVar(&cli.HttpsAddr, "https-addr", ":8443", "Public HTTPS address")
	fs.BoolVar(&cli.Insecure, "insecure", false, "Run without HTTPS. not recomended, as WebRTC support requires HTTPS")
	cli.DataDirFlag(fs, &cli.TLSCertPath, "tls-cert", filepath.Join("tls", "tls.crt"), "Path to TLS certificate")
	cli.DataDirFlag(fs, &cli.TLSKeyPath, "tls-key", filepath.Join("tls", "tls.crt"), "Path to TLS key")
	fs.StringVar(&cli.SigningKeyPath, "signing-key", "", "Path to signing key for pushing OTA updates to the app")
	cli.DataDirFlag(fs, &cli.DBPath, "db-path", "db.sqlite", "path to sqlite database file")
	fs.StringVar(&cli.AdminAccount, "admin-account", "", "ethereum account that administrates this aquareum node")
	fs.StringVar(&cli.FirebaseServiceAccount, "firebase-service-account", "", "JSON string of a firebase service account key")
	fs.StringVar(&cli.GitLabURL, "gitlab-url", "https://git.aquareum.tv/api/v4/projects/1", "gitlab url for generating download links")
	cli.DataDirFlag(fs, &cli.EthKeystorePath, "eth-keystore-path", "keystore", "path to ethereum keystore")
	fs.StringVar(&cli.EthAccountAddr, "eth-account-addr", "", "ethereum account address to use (if keystore contains more than one)")
	fs.StringVar(&cli.EthPassword, "eth-password", "", "password for encrypting keystore")
	fs.StringVar(&cli.TAURL, "ta-url", "http://timestamp.digicert.com", "timestamp authority server for signing")
	version := fs.Bool("version", false, "print version and exit")

	if runtime.GOOS == "linux" {
		fs.IntVar(&cli.MistAdminPort, "mist-admin-port", 14242, "MistServer admin port (internal use only)")
		fs.IntVar(&cli.MistRTMPPort, "mist-rtmp-port", 11935, "MistServer RTMP port (internal use only)")
		fs.IntVar(&cli.MistHTTPPort, "mist-http-port", 18080, "MistServer HTTP port (internal use only)")
	}

	cli.Parse(
		fs, os.Args[1:],
	)

	ctx := context.Background()

	log.Log(ctx,
		"aquareum",
		"version", build.Version,
		"buildTime", build.BuildTimeStr(),
		"uuid", build.UUID,
		"runtime.GOOS", runtime.GOOS,
		"runtime.GOARCH", runtime.GOARCH)
	if *version {
		return nil
	}

	err = os.MkdirAll(cli.DataDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating aquareum dir at %s:%w", cli.DataDir, err)
	}
	schema, err := v0.MakeV0Schema()
	if err != nil {
		return err
	}
	signer, err := eip712.MakeEIP712Signer(ctx, &eip712.EIP712SignerOptions{
		Schema:              schema,
		EthKeystorePath:     cli.EthKeystorePath,
		EthAccountAddr:      cli.EthAccountAddr,
		EthKeystorePassword: cli.EthPassword,
	})
	if err != nil {
		return err
	}
	mm, err := media.MakeMediaManager(ctx, &cli, signer)
	if err != nil {
		return err
	}
	mod, err := model.MakeDB(cli.DBPath)
	if err != nil {
		return err
	}
	var noter notifications.FirebaseNotifier
	if cli.FirebaseServiceAccount != "" {
		noter, err = notifications.MakeFirebaseNotifier(ctx, cli.FirebaseServiceAccount)
		if err != nil {
			return err
		}
	}
	a, err := api.MakeAquareumAPI(&cli, mod, signer, noter, mm)
	if err != nil {
		return err
	}

	group, ctx := TimeoutGroupWithContext(ctx)
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

	for _, job := range platformJobs {
		group.Go(func() error {
			return job(ctx, &cli)
		})
	}

	return group.Wait()
}

func handleSignals(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT)
	for {
		select {
		case s := <-c:
			if s == syscall.SIGABRT {
				pprof.Lookup("goroutine").WriteTo(os.Stderr, 2)
			}
			log.Log(ctx, "caught signal, attempting clean shutdown", "signal", s)
			return fmt.Errorf("caught signal=%v", s)
		case <-ctx.Done():
			return nil
		}
	}
}
