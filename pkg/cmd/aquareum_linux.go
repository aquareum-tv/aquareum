//go:build linux

package cmd

import (
	"context"

	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/proc"
)

func runMist(ctx context.Context, cli *config.CLI) error {
	if cli.NoMist {
		<-ctx.Done()
		return nil
	}
	return proc.RunMistServer(ctx, cli)
}

func Start(build *config.BuildFlags) error {
	return start(build, []jobFunc{runMist})
}
