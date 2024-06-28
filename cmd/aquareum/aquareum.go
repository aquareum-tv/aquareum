package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"aquareum.tv/aquareum/packages/app"
	"github.com/adrg/xdg"
)

var Version = "unknown"

func main() {
	err := start()
	if err != nil {
		log.Fatal(err)
	}
}

func start() error {
	if xdg.Home == "/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		os.Setenv("HOME", home)
		xdg.Reload()
	}
	if xdg.Home == "/" {
		return fmt.Errorf("couldn't find users home directory")
	}
	tlsCrtFile, err := xdg.ConfigFile("aquareum/tls/tls.crt")
	if err != nil {
		log.Fatal(err)
	}
	tlsKeyFile, err := xdg.ConfigFile("aquareum/tls/tls.key")
	if err != nil {
		log.Fatal(err)
	}
	files, err := app.Files()
	if err != nil {
		return err
	}
	http.Handle("/", http.FileServer(http.FS(files)))
	err = http.ListenAndServeTLS(":443", tlsCrtFile, tlsKeyFile, nil)
	if err != nil {
		return err
	}
	return nil
}
