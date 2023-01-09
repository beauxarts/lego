package cli

import (
	"github.com/beauxarts/lego/chapter_paragraph"
	"github.com/boggydigital/nod"
	"net/http"
	"net/url"
	"os"
)

func VoicesHandler(u *url.URL) error {
	q := u.Query()

	provider := q.Get("provider")

	key := q.Get("key-value")
	if key == "" {
		//attempt to get the key from a file, if specified
		keyFilename := q.Get("key-filename")
		if keyBytes, err := os.ReadFile(keyFilename); err == nil {
			key = string(keyBytes)
		}
	}
	locale := q.Get("locale")

	return Voices(provider, key, locale)
}

func Voices(provider, key, locale string) error {

	va := nod.Begin("available voices for the selected provider:")
	defer va.End()

	var szr *chapter_paragraph.Synthesizer
	var err error

	switch provider {
	case "gcp":
		szr, err = chapter_paragraph.NewGCPSynthesizer(http.DefaultClient, nil, key, "", false)
	case "say":
		szr, err = chapter_paragraph.NewSaySynthesizer(nil, "", false)
	}

	if err != nil {
		return va.EndWithError(err)
	}

	voices, err := szr.Voices(locale)
	if err != nil {
		return va.EndWithError(err)
	}

	for _, vs := range voices {
		if vs == "" {
			continue
		}
		v := nod.Begin("- " + vs)
		v.End()
	}

	va.EndWithResult("done")

	return nil
}
