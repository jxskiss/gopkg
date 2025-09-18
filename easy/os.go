package easy

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/jxskiss/gopkg/v2/utils/strutil"
)

// ReadFileLines reads all lines from a file.
// It automatically removes BOM bytes from the head of the file content if exists.
func ReadFileLines(filename string) ([]string, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	var lines []string
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(lines) == 0 {
			line = strutil.TrimBOM(line)
		}
		lines = append(lines, string(line))
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
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
