package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/proc"

	"aquareum.tv/aquareum/pkg/api"
	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/model"
	"github.com/adrg/xdg"
	"github.com/peterbourgon/ff/v3"
	"golang.org/x/sync/errgroup"
)

// parse the CLI and fire up an aquareum node!
func Start(build *config.BuildFlags) error {
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
	fs.StringVar(&cli.HttpsAddr, "https-addr", ":8443", "Public HTTPS address")
	fs.BoolVar(&cli.Insecure, "insecure", false, "Run without HTTPS. not recomended, as WebRTC support requires HTTPS")
	fs.StringVar(&cli.TLSCertPath, "tls-cert", tlsCertFile, "Path to TLS certificate")
	fs.StringVar(&cli.TLSKeyPath, "tls-key", tlsKeyFile, "Path to TLS key")
	fs.StringVar(&cli.SigningKeyPath, "signing-key", "", "Path to signing key for pushing OTA updates to the app")
	fs.StringVar(&cli.DBPath, "db-path", dbFile, "path to sqlite database file")
	fs.StringVar(&cli.AdminSecret, "admin-secret", "", "secret admin token (to be replaced soon)")

	ff.Parse(
		fs, os.Args[1:],
		ff.WithEnvVarPrefix("AQ"),
		ff.WithEnvVarSplit(","),
	)

	log.Log(context.Background(),
		"starting aquareum",
		"version", build.Version,
		"buildTime", build.BuildTimeStr(),
		"uuid", build.UUID)

	mod, err := model.MakeDB(cli.DBPath)
	if err != nil {
		return err
	}
	a, err := api.MakeAquareumAPI(&cli, mod)
	if err != nil {
		return err
	}

	group, ctx := errgroup.WithContext(context.Background())
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
		return proc.RunMistServer(ctx)
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
