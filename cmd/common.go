package cmd

import (
	"bufio"
	"github.com/thoughtrealm/bumblebee/helpers"
	"os"
)

const BACKUP_FILE_METADATA_NAME = "backup-details"

type BackupDetailsMetadata struct {
	Profiles []*helpers.Profile
}

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
