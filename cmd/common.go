package cmd

import (
	"bufio"
	"os"
)

func getDescriptorPaths(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var paths []string
	for scanner.Scan() {
		paths = append(paths, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return paths, nil
}
