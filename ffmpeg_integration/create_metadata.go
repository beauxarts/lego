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
	MetadataAlbumPrefix          = "album="
	MetadataDescriptionPrefix    = "description="
	MetadataGenrePrefix          = "genre="
	MetadataYearPrefix           = "year="
	MetadataCopyrightPrefix      = "copyright="
	MetadataPerformerPrefix      = "performer="
	MetadataTimebaseDefaultValue = "TIMEBASE=1/1000"
	MetadataStartPrefix          = "START="
	MetadataEndPrefix            = "END="
	MetadataFilePrefix           = "file="
	MetadataFilename             = "_ffmpegmetadata.txt"
)

func CreateMetadata(
	filename string,
	metadata map[string]string,
	chapters []string,
	chaptersDuration map[string]int64) error {

	sb := strings.Builder{}

	//https://ffmpeg.org/ffmpeg-all.html#Metadata-1
	sb.WriteString(MetadataHeader + "\n")
	// write metadata
	for property, value := range metadata {
		switch property {
		case "title":
			sb.WriteString(MetadataTitlePrefix + value + "\n")
			sb.WriteString(MetadataAlbumPrefix + value + "\n")
		case "authors":
			sb.WriteString(MetadataArtistPrefix + value + "\n")
		case "date-created":
			sb.WriteString(MetadataYearPrefix + value + "\n")
		case "genres":
			sb.WriteString(MetadataGenrePrefix + value + "\n")
		case "copyright-holders":
			sb.WriteString(MetadataCopyrightPrefix + value + "\n")
		case "description":
			sb.WriteString(MetadataDescriptionPrefix + value + "\n")
		case "readers":
			sb.WriteString(MetadataPerformerPrefix + value + "\n")
		default:
			sb.WriteString(property + "=" + value + "\n")
		}
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
