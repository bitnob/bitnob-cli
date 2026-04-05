package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bitnob/bitnob-cli/internal/app"
	"github.com/bitnob/bitnob-cli/internal/cli"
	"github.com/bitnob/bitnob-cli/internal/version"
)

// Build variables set by ldflags
var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

func main() {
	ctx := context.Background()

	buildInfo := version.Info{
		Version: buildVersion,
		Commit:  buildCommit,
		Date:    buildDate,
	}

	application, err := app.New(ctx, app.Options{
		Version:    buildInfo,
		ConfigPath: os.Getenv("BITNOB_CONFIG_PATH"),
		StateDir:   os.Getenv("BITNOB_STATE_DIR"),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, warning := range application.StartupWarnings {
		fmt.Fprintln(os.Stderr, warning)
	}

	if err := cli.Run(ctx, application, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
