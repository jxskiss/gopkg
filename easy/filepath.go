package easy

import (
	"os"
	"path/filepath"
	"strings"
)

// Glob adds double-star support to the std library's [path/filepath.Glob].
// It's useful when your pattern might have double-stars.
func Glob(pattern string) (matches []string, err error) {
	if !strings.Contains(pattern, "**") {
		// Pass-through to std lib if no double-star.
		return filepath.Glob(pattern)
	}
	parts := strings.Split(pattern, "**")
	return globParts(parts).Expand()
}

type globParts []string

func (globs globParts) Expand() (matches []string, err error) {
	matches = []string{""}
	for i, glob := range globs {
		isLast := i == len(globs)-1
		var hits []string
		var seen = make(map[string]bool)
		for _, match := range matches {
			paths, err := filepath.Glob(match + glob)
			if err != nil {
				return nil, err
			}
			for _, path := range paths {
				err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if isLast || info.IsDir() {
						if _, ok := seen[path]; !ok {
							hits = append(hits, path)
							seen[path] = true
						}
					}
					return nil
				})
				if err != nil {
					return nil, err
				}
			}
		}
		matches = hits
	}

	if len(globs) == 0 && len(matches) > 0 && matches[0] == "" {
		matches = matches[1:]
	}

	return matches, nil
}

// CreateNonExistingFolder checks whether a directory exists,
// the directory will be created by calling `os.MkdirAll(path, perm)`
// if it does not exist.
func CreateNonExistingFolder(path string, perm os.FileMode) error {
	if perm == 0 {
		perm = 0o700
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, perm)
	} else if err != nil {
		return err
	}
	return nil
}

// WriteFile writes data to the named file, creating it if necessary.
// If the file does not exist, WriteFile creates it with permissions perm (before umask);
// otherwise WriteFile truncates it before writing, without changing permissions.
//
// If creates the directory if it does not exist instead of reporting an error.
func WriteFile(name string, data []byte, perm os.FileMode) error {
	dirPerm := getDirectoryPermFromFilePerm(perm)
	err := CreateNonExistingFolder(filepath.Dir(name), dirPerm)
	if err != nil {
		return err
	}
	return os.WriteFile(name, data, perm)
}

func getDirectoryPermFromFilePerm(filePerm os.FileMode) os.FileMode {
	var dirPerm os.FileMode = 0o700
	if filePerm&0o060 > 0 {
		dirPerm |= (filePerm & 0o070) | 0o010
	}
	if filePerm&0o006 > 0 {
		dirPerm |= (filePerm & 0o007) | 0x001
	}
	return dirPerm
}
