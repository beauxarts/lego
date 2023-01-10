package cli

import (
	"errors"
	"github.com/beauxarts/divido"
	"github.com/boggydigital/nod"
	"math"
	"net/url"
	"os"
	"strconv"
)

const (
	defaultSayCostPerMillionChars = 0.0
	defaultGCPCostPerMillionChars = 16.0
)

func InfoHandler(u *url.URL) error {
	q := u.Query()

	filename := q.Get("filename")
	provider := q.Get("provider")

	cpmc := defaultSayCostPerMillionChars
	switch provider {
	case "gcp":
		cpmc = defaultGCPCostPerMillionChars
	case "say":
		//do nothing
	}

	cpmcs := q.Get("cost-per-million-characters")
	if cpmci, err := strconv.ParseFloat(cpmcs, 64); err == nil {
		cpmc = cpmci
	}

	return Info(filename, provider, cpmc)
}

func Info(filename, provider string, costPerMillionChars float64) error {
	ia := nod.Begin("document info:")
	defer ia.End()

	cea := nod.Begin(" estimating synthesis cost (%s)...", provider)
	defer cea.End()

	if stat, err := os.Stat(filename); err == nil {
		estCost := float64(stat.Size()) * costPerMillionChars / math.Pow(10, 6)
		cea.EndWithResult("~$%.2f (at $%.2f/1M chars)", estCost, costPerMillionChars)
	} else {
		return cea.EndWithError(errors.New("input file not found: " + filename))
	}

	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return ia.EndWithError(err)
	}

	td := divido.NewTextDocument(file)

	cla := nod.Begin("chapter info:")
	defer cla.End()

	for _, ct := range td.ChapterTitles() {
		cta := nod.Begin(" \"%s\":", ct)

		ln := 0
		maxln := 0
		paragraphs := td.ChapterParagraphs(ct)
		for _, p := range paragraphs {
			lp := len(p)
			ln += lp
			if lp > maxln {
				maxln = lp
			}
		}

		cia := nod.Begin("")
		cia.EndWithResult("- %d paragraphs, longest: %d chars, total length: %d chars", len(paragraphs), maxln, ln)

		cta.End()
	}

	ia.EndWithResult("done")

	return nil
}
