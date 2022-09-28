package h

import (
	"fmt"
	"os"
)

func Die(msg string, v ...any) {
	fmt.Printf("error: %s\n", fmt.Sprintf(msg, v...))
	os.Exit(1)
}
