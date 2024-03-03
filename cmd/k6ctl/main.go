package main

import (
	"github.com/alecthomas/kong"
)

func main() {
	cli := &CLI{}
	cliCtx := kong.Parse(cli)
	err := cliCtx.Run()
	cliCtx.FatalIfErrorf(err)
}
