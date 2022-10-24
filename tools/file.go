package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func ReadContentAsList(path string) ([]string, error) {
	var content []string
	file, err := os.Open(path)
	if err != nil {
		return content, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}

	return content, scanner.Err()
}

func RemoveSubFileFolder(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		log("deleting: %+v", filepath.Join(dir, name))
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func log(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}
