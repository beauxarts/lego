package cli

import (
	"github.com/beauxarts/polyglot/gcp"
	"github.com/boggydigital/nod"
	"net/http"
	"net/url"
	"os"
)

func LanguagesHandler(u *url.URL) error {
	q := u.Query()

	target := q.Get("target")
	key := q.Get("key-value")
	if key == "" {
		//attempt to get the key from a file, if specified
		keyFilename := q.Get("key-filename")
		if keyBytes, err := os.ReadFile(keyFilename); err == nil {
			key = string(keyBytes)
		}
	}

	return Languages(target, key)
}

func Languages(target, key string) error {
	la := nod.Begin("languages available for translations:")
	defer la.End()

	translator, err := gcp.NewTranslator(http.DefaultClient, gcp.NeuralMachineTranslation, key)
	if err != nil {
		return la.EndWithError(err)
	}

	languages, err := translator.Languages(target)
	if err != nil {
		return la.EndWithError(err)
	}

	for lang, name := range languages {
		v := nod.Begin("- %s: %s", lang, name)
		v.End()
	}

	la.EndWithResult("done")

	return nil
}
