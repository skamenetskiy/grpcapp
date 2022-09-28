package h

import "os"

func DirInfo(name string) (bool, bool) {
	dir, err := os.ReadDir(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, true
		}
		Die("failed to read dir: %s", err)
	}
	if len(dir) != 0 {
		return true, false
	}
	return true, true
}
