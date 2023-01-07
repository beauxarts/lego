package ffmpeg_integration

import (
	"bufio"
	"os"
	"strings"
)

func ExtractFiles(filename string) ([]string, error) {
	metadataFile, err := os.Open(filename)
	defer metadataFile.Close()
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)

	scanner := bufio.NewScanner(metadataFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, MetadataFilePrefix) {
			files = append(files, strings.TrimPrefix(line, MetadataFilePrefix))
		}
	}

	return files, nil
}
