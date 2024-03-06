package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
)

type CLIVersion struct {
	Output io.Writer `kong:"-"`
}

func (c *CLIVersion) BeforeApply() error {
	if c.Output == nil {
		c.Output = os.Stderr
	}

	return nil
}

func (c *CLIVersion) Run() error {
	_, err := fmt.Fprintln(c.Output, version())
	return err
}

func version() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	var (
		goVersion string
		revision  string
		modified  string
		buildTime string
	)
	goVersion = buildInfo.GoVersion
	for _, s := range buildInfo.Settings {
		if s.Value == "" {
			continue
		}

		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value
		case "vcs.time":
			buildTime = s.Value
		}
	}

	// in Go install mode, this is a known issue that vcs information will not be available.
	// ref: https://github.com/golang/go/issues/51279
	// Fallback to use module version and stop here as vcs information is incomplete.
	if revision == "" {
		if buildInfo.Main.Version != "" {
			// fallback to use module version (legacy usage)
			revision = buildInfo.Main.Version
		}

		// stop here as vcs information is incomplete
		return fmt.Sprintf("%s revision: %s", goVersion, revision)
	}

	return fmt.Sprintf("%s revision: %s modified: %s buildTime: %s", goVersion, revision, modified, buildTime)
}
