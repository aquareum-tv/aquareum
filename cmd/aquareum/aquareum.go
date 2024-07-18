package main

import (
	"context"

	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/log"

	"aquareum.tv/aquareum/pkg/cmd"
)

var Version = "unknown"
var BuildTime int64 = 0

func main() {
	err := cmd.Start(&config.BuildFlags{
		Version:   Version,
		BuildTime: BuildTime,
	})
	if err != nil {
		log.Log(context.Background(), "exited uncleanly", "error", err)
	}
}
