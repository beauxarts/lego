package chapter_paragraph

import (
	"fmt"
	gti "github.com/beauxarts/google_tts_integration"
)

const (
	filenamePaddedDigits = "09"
	listExt              = ".txt"
	filesSuffix          = "_files"
	ffmpegSuffix         = "_ffmpeg"
	bookPrefix           = "_book"
)

func RelChapterTitlesFilename() string {
	return chapterTitlesFilename
}

func RelChapterFilename(chapter int) string {
	return fmt.Sprintf("%"+filenamePaddedDigits+"d"+gti.DefaultEncodingExt, chapter)
}

func RelChapterFfmpegOutputFilename(chapter int) string {
	return fmt.Sprintf("%"+filenamePaddedDigits+"d"+ffmpegSuffix+listExt, chapter)
}

func RelChapterTitleFilename(chapter int) string {
	return RelChapterParagraphFilename(chapter, 0)
}

func RelChapterParagraphFilename(chapter, paragraph int) string {
	return fmt.Sprintf(
		"%"+filenamePaddedDigits+"d-%"+filenamePaddedDigits+"d"+gti.DefaultEncodingExt,
		chapter,
		paragraph)
}

func RelChapterFilesFilename(chapter int) string {
	return fmt.Sprintf("%"+filenamePaddedDigits+"d"+filesSuffix+listExt, chapter)
}

func RelBookFilesFilename() string {
	return bookPrefix + filesSuffix + listExt
}
