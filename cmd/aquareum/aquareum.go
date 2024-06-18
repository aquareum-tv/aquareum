package main

import (
	"fmt"
	"log"
	"net/http"

	"aquareum.tv/aquareum/packages/app"
)

var Version = "unknown"

func main() {
	files, err := app.Files()
	if err != nil {
		panic(err)
	}
	http.Handle("/", http.FileServer(http.FS(files)))
	fmt.Printf("aquareum version %s\n", Version)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
