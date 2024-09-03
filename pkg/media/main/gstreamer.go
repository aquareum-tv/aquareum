package main

import (
	"context"
	"fmt"
	"os"

	"aquareum.tv/aquareum/pkg/media"
)

func Main(fpath string) error {
	ifile, err := os.Open(fpath)
	if err != nil {
		return err
	}
	ofile, err := os.CreateTemp("", "*.mkv")
	if err != nil {
		return err
	}
	fmt.Println(ofile.Name())
	err = media.NormalizeAudio(context.Background(), ifile, ofile)
	if err != nil {
		return err
	}
	panic(ofile.Name())
}

func main() {
	err := Main(os.Args[1])
	if err != nil {
		panic(err)
	}
}
