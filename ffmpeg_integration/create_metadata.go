package ffmpeg_integration

import (
	"github.com/beauxarts/lego/chapter_paragraph"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	MetadataHeader               = ";FFMETADATA1"
	MetadataChapterSection       = "[CHAPTER]"
	MetadataTitlePrefix          = "title="
	MetadataArtistPrefix         = "artist="
	MetadataTimebaseDefaultValue = "TIMEBASE=1/1000"
	MetadataStartPrefix          = "START="
	MetadataEndPrefix            = "END="
	MetadataFilePrefix           = "file="
	MetadataFilename             = "_ffmpegmetadata.txt"
)

func CreateMetadata(filename, title, author string, chapters []string, chaptersDuration map[string]int64) error {

	sb := strings.Builder{}

	//https://ffmpeg.org/ffmpeg-all.html#Metadata-1
	sb.WriteString(MetadataHeader + "\n")
	if title != "" {
		sb.WriteString(MetadataTitlePrefix + title + "\n")
	}
	if author != "" {
		sb.WriteString(MetadataArtistPrefix + author + "\n")
	}
	sb.WriteString("\n")

	var currentOffset int64 = 0

	for ci, ct := range chapters {
		sb.WriteString(MetadataChapterSection + "\n")
		sb.WriteString(MetadataTimebaseDefaultValue + "\n")
		sb.WriteString(MetadataStartPrefix + strconv.FormatInt(currentOffset, 10) + "\n")
		currentOffset += chaptersDuration[ct]
		sb.WriteString(MetadataEndPrefix + strconv.FormatInt(currentOffset, 10) + "\n")
		sb.WriteString(MetadataTitlePrefix + ct + "\n")
		sb.WriteString(MetadataFilePrefix + chapter_paragraph.RelChapterFilename(ci+1) + "\n")
		sb.WriteString("\n")
	}

	metadataFile, err := os.Create(filename)
	defer metadataFile.Close()
	if err != nil {
		return err
	}

	_, err = io.WriteString(metadataFile, sb.String())
	return err
}
