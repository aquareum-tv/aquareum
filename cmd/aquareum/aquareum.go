package main

import (
	"fmt"
	"log"
	"net/http"

	"aquareum.tv/aquareum/packages/app"
)

func main() {
	fmt.Println("hello world")
	files, err := app.Files()
	if err != nil {
		panic(err)
	}
	http.Handle("/", http.FileServer(http.FS(files)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
