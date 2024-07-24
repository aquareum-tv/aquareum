package main

import (
	"context"
	"strconv"

	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/log"

	"aquareum.tv/aquareum/pkg/cmd"
)

import "C"

var Version = "unknown"
var BuildTime = "0"
var UUID = ""

//export AquareumMain
func AquareumMain() {
	i, err := strconv.ParseInt(BuildTime, 10, 64)
	if err != nil {
		panic(err)
	}
	err = cmd.Start(&config.BuildFlags{
		Version:   Version,
		BuildTime: i,
		UUID:      UUID,
	})
	if err != nil {
		log.Log(context.Background(), "exited uncleanly", "error", err)
	}
}

func main() {
	AquareumMain()
}
