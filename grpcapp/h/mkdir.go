package h

import (
	"os"
)

func Mkdir(name string) {
	const dirPerm = 0775
	if err := os.MkdirAll(name, dirPerm); err != nil {
		Die("failed to create directory: %s", err)
	}
}
