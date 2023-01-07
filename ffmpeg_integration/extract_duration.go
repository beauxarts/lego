package ffmpeg_integration

import (
	"bufio"
	"os"
	"strings"
	"time"
)

const (
	sizePrefix = "size="
	timePrefix = "time="
)

var zeroDate = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)

func ExtractChapterDuration(filename string) (int64, error) {

	outputFile, err := os.Open(filename)
	defer outputFile.Close()
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(outputFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, sizePrefix) {
			parts := strings.Split(line, " ")
			for _, p := range parts {
				if strings.HasPrefix(p, timePrefix) {
					ts := strings.TrimPrefix(p, timePrefix)
					if td, err := time.Parse("15:04:5.00", ts); err == nil {
						dur := td.Sub(zeroDate)
						if dur.Milliseconds() > 0 {
							return dur.Milliseconds(), nil
						}
					} else {
						return 0, err
					}
				}
			}
		}
	}

	return 0, nil
}
