package help

import (
	"fmt"
)

func Run(_ []string) {
	fmt.Println(`usage: grpcapp {command} [...options]

commands:
	create {name} - create new application
	generate      - generate proto
	help          - print help information`)
}
