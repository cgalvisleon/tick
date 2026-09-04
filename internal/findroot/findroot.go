// Package findroot locates the nearest .tick directory by walking up from
// the current working directory, the same way git locates .git.
package findroot

import (
	"fmt"
	"os"
	"path/filepath"
)

const DirName = ".tick"

/**
* Find: Walks up from the current working directory looking for a .tick
* directory. Returns the directory containing it (the project root).
* @return string, error
**/
func Find() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, DirName)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a tick project (or any parent up to /): run 'tick init' first")
		}
		dir = parent
	}
}

/**
* DotDir: Returns the .tick directory path for the given project root.
* @param root string
* @return string
**/
func DotDir(root string) string {
	return filepath.Join(root, DirName)
}
