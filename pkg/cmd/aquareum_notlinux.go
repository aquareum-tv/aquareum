//go:build !linux

package cmd

import "aquareum.tv/aquareum/pkg/config"

func Start(build *config.BuildFlags) error {
	return start(build, []jobFunc{})
}
