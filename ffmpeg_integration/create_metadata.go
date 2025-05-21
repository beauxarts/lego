package ffmpeg_integration

import (
	"io"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
)

const (
	MetadataHeader               = ";FFMETADATA1"
	MetadataChapterSection       = "[CHAPTER]"
	MetadataTitlePrefix          = "title="
	MetadataAuthorPrefix         = "author="
	MetadataArtistPrefix         = "artist="
	MetadataAlbumPrefix          = "album="
	MetadataTimebaseDefaultValue = "TIMEBASE=1/1000"
	MetadataStartPrefix          = "START="
	MetadataEndPrefix            = "END="
	MetadataFilePrefix           = "file="
	MetadataFilename             = "_ffmpegmetadata.txt"
)

func CreateMetadata(
	filename string,
	title, author string,
	chapterFilenameTitle map[string]string,
	chaptersFileDuration map[string]int64) error {

	sb := strings.Builder{}

	//https://ffmpeg.org/ffmpeg-all.html#Metadata-1
	sb.WriteString(MetadataHeader + "\n")
	// write metadata
	if title != "" {
		sb.WriteString(MetadataTitlePrefix + title + "\n")
		sb.WriteString(MetadataAlbumPrefix + title + "\n")
	}
	if author != "" {
		sb.WriteString(MetadataArtistPrefix + author + "\n")
		sb.WriteString(MetadataAuthorPrefix + author + "\n")
	}
	sb.WriteString("\n")

	var currentOffset int64 = 0

	cfns := maps.Keys(chapterFilenameTitle)
	sortedCfns := slices.Sorted(cfns)

	for _, cfn := range sortedCfns {
		sb.WriteString(MetadataChapterSection + "\n")
		sb.WriteString(MetadataTimebaseDefaultValue + "\n")
		sb.WriteString(MetadataStartPrefix + strconv.FormatInt(currentOffset, 10) + "\n")
		currentOffset += chaptersFileDuration[cfn]
		sb.WriteString(MetadataEndPrefix + strconv.FormatInt(currentOffset, 10) + "\n")
		sb.WriteString(MetadataTitlePrefix + chapterFilenameTitle[cfn] + "\n")
		sb.WriteString(MetadataFilePrefix + cfn + "\n")
		sb.WriteString("\n")
	}

	metadataFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer metadataFile.Close()

	_, err = io.WriteString(metadataFile, sb.String())
	return err
}
