package main

import (
	"os"

	"github.com/qutty/pushserver/cli"
)

func main() {
	if len(os.Args) > 1 {
		cli.Run(os.Args[1:])
		return
	}

	// Default behavior - start server
	cli.StartServer()
}
