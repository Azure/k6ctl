package main

type CLI struct {
	Verbose bool `short:"v" long:"verbose" description:"Show verbose debug information"`

	Run     CLIRun     `cmd:"run" help:"Run a k6 task"`
	Version CLIVersion `cmd:"version" help:"Show the k6ctl version"`
}
