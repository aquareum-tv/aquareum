package cmd

import (
	"context"
	"crypto"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"syscall"

	"aquareum.tv/aquareum/pkg/crypto/signers"
	"aquareum.tv/aquareum/pkg/crypto/signers/eip712"
	"aquareum.tv/aquareum/pkg/log"
	"aquareum.tv/aquareum/pkg/media"
	"aquareum.tv/aquareum/pkg/notifications"
	v0 "aquareum.tv/aquareum/pkg/schema/v0"
	"golang.org/x/term"

	"aquareum.tv/aquareum/pkg/api"
	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/model"
	"github.com/ThalesGroup/crypto11"
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
	cli.DataDirFlag(fs, &cli.TLSKeyPath, "tls-key", filepath.Join("tls", "tls.key"), "Path to TLS key")
	fs.StringVar(&cli.SigningKeyPath, "signing-key", "", "Path to signing key for pushing OTA updates to the app")
	cli.DataDirFlag(fs, &cli.DBPath, "db-path", "db.sqlite", "path to sqlite database file")
	fs.StringVar(&cli.AdminAccount, "admin-account", "", "ethereum account that administrates this aquareum node")
	fs.StringVar(&cli.FirebaseServiceAccount, "firebase-service-account", "", "JSON string of a firebase service account key")
	fs.StringVar(&cli.GitLabURL, "gitlab-url", "https://git.aquareum.tv/api/v4/projects/1", "gitlab url for generating download links")
	cli.DataDirFlag(fs, &cli.EthKeystorePath, "eth-keystore-path", "keystore", "path to ethereum keystore")
	fs.StringVar(&cli.EthAccountAddr, "eth-account-addr", "", "ethereum account address to use (if keystore contains more than one)")
	fs.StringVar(&cli.EthPassword, "eth-password", "", "password for encrypting keystore")
	fs.StringVar(&cli.TAURL, "ta-url", "http://timestamp.digicert.com", "timestamp authority server for signing")
	fs.StringVar(&cli.PKCS11ModulePath, "pkcs11-module-path", "", "path to a PKCS11 module for HSM signing, for example /usr/lib/x86_64-linux-gnu/opensc-pkcs11.so")
	fs.StringVar(&cli.PKCS11Pin, "pkcs11-pin", "", "PIN for logging into PKCS11 token. if not provided, will be prompted interactively")
	fs.StringVar(&cli.PKCS11TokenSlot, "pkcs11-token-slot", "", "slot number of PKCS11 token (only use one of slot, label, or serial)")
	fs.StringVar(&cli.PKCS11TokenLabel, "pkcs11-token-label", "", "label of PKCS11 token (only use one of slot, label, or serial)")
	fs.StringVar(&cli.PKCS11TokenSerial, "pkcs11-token-serial", "", "serial number of PKCS11 token (only use one of slot, label, or serial)")
	fs.StringVar(&cli.PKCS11KeypairLabel, "pkcs11-keypair-label", "", "label of signing keypair on PKCS11 token")
	fs.StringVar(&cli.PKCS11KeypairID, "pkcs11-keypair-id", "", "id of signing keypair on PKCS11 token")
	fs.StringVar(&cli.StreamerName, "streamer-name", "", "name of the person streaming from this aquareum node")
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
	eip712signer, err := eip712.MakeEIP712Signer(ctx, &eip712.EIP712SignerOptions{
		Schema:              schema,
		EthKeystorePath:     cli.EthKeystorePath,
		EthAccountAddr:      cli.EthAccountAddr,
		EthKeystorePassword: cli.EthPassword,
	})
	if err != nil {
		return err
	}
	var signer crypto.Signer = eip712signer
	if cli.PKCS11ModulePath != "" {
		conf := &crypto11.Config{
			Path: cli.PKCS11ModulePath,
		}
		count := 0
		for _, val := range []string{cli.PKCS11TokenSlot, cli.PKCS11TokenLabel, cli.PKCS11TokenSerial} {
			if val != "" {
				count += 1
			}
		}
		if count != 1 {
			return fmt.Errorf("need exactly one of pkcs11-token-slot, pkcs11-token-label, or pkcs11-token-serial (got %d)", count)
		}
		if cli.PKCS11TokenSlot != "" {
			num, err := strconv.ParseInt(cli.PKCS11TokenSlot, 10, 16)
			if err != nil {
				return fmt.Errorf("error parsing pkcs11-slot: %w", err)
			}
			numint := int(num)
			// why does crypto11 want this as a reference? odd.
			conf.SlotNumber = &numint
		}
		if cli.PKCS11TokenLabel != "" {
			conf.TokenLabel = cli.PKCS11TokenLabel
		}
		if cli.PKCS11TokenSerial != "" {
			conf.TokenSerial = cli.PKCS11TokenSerial
		}
		pin := cli.PKCS11Pin
		if pin == "" {
			fmt.Printf("Please enter PKCS11 PIN: ")
			password, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println("")
			if err != nil {
				return fmt.Errorf("error reading PKCS11 password: %w", err)
			}
			pin = string(password)
		}
		conf.Pin = pin

		sc, err := crypto11.Configure(conf)
		if err != nil {
			return fmt.Errorf("error initalizing PKCS11 HSM: %w", err)
		}
		var id []byte = nil
		var label []byte = nil
		if cli.PKCS11KeypairID != "" {
			num, err := strconv.ParseInt(cli.PKCS11KeypairID, 10, 8)
			if err != nil {
				return fmt.Errorf("error parsing pkcs11-keypair-id: %w", err)
			}
			id = []byte{byte(num)}
		}
		if cli.PKCS11KeypairLabel != "" {
			label = []byte(cli.PKCS11KeypairLabel)
		}
		hwsigner, err := sc.FindKeyPair(id, label)
		if err != nil {
			return fmt.Errorf("error finding keypair on PKCS11 token: %w", err)
		}
		if hwsigner == nil {
			return fmt.Errorf("keypair on token not found (tried id='%s' label='%s')", cli.PKCS11KeypairID, cli.PKCS11KeypairLabel)
		}
		addr, err := signers.HexAddrFromSigner(hwsigner)
		if err != nil {
			return fmt.Errorf("error getting ethereum address for hardware keypair: %w", err)
		}
		log.Log(ctx, "successfully initialized hardware signer", "address", addr)
		signer = hwsigner
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
	a, err := api.MakeAquareumAPI(&cli, mod, eip712signer, noter, mm)
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
