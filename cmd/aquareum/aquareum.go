package main

import (
	"context"

	"aquareum.tv/aquareum/pkg/log"

	"aquareum.tv/aquareum/pkg/cmd"
)

var Version = "unknown"

func main() {
	err := cmd.Start(&cmd.BuildFlags{
		Version: Version,
	})
	if err != nil {
		log.Log(context.Background(), "exited uncleanly", "error", err)
	}
}
