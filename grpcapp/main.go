package main

import (
	"os"

	"github.com/skamenetskiy/grpcapp/grpcapp/create"
	"github.com/skamenetskiy/grpcapp/grpcapp/generate"
	"github.com/skamenetskiy/grpcapp/grpcapp/help"
)

func main() {
	var (
		args    = os.Args[1:]
		command string
	)
	if len(args) == 0 {
		command = "help"
	} else {
		command = args[0]
		args = args[1:]
	}
	var cmd func([]string)
	switch command {
	case "create":
		cmd = create.Run
	case "generate":
		cmd = generate.Run
	case "help":
		cmd = help.Run
	default:
		cmd = help.Run
	}
	cmd(args)
}
