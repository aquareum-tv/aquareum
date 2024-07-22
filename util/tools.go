//go:build tools

package main

import (
	_ "gitlab.com/gitlab-org/release-cli/cmd/release-cli"
)

// This file just exists so we can `go install` stuff with pinned versions later
