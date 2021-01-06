package library

import (
	"encoding/json"
	"os"
)

type keywords struct {
	keywords map[string]struct{}
	file     string
}

func loadKeywords(file string) (*keywords, error) {
	// make empty return value
	ks := &keywords{make(map[string]struct{}), file}

	// load keywords file
	f, err := os.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			// make default file
			ks.save()
			return ks, nil
		}
		return nil, err
	}
	defer f.Close()
	// decode file as keyword array
	s := make([]string, 0)
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}
	// put keywords to return value (convert slice to map)
	for _, ss := range s {
		ks.keywords[ss] = struct{}{}
	}
	return ks, nil
}

func (k *keywords) save() error {
	ks := k.keywordsSlice()
	f, err := os.Create(k.file)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(ks); err != nil {
		return err
	}
	return nil
}

func (k *keywords) add(keyword string) error {
	k.keywords[keyword] = struct{}{}
	return k.save()
}

func (k *keywords) remove(keyword string) error {
	delete(k.keywords, keyword)
	return k.save()
}

func (k *keywords) keywordsSlice() []string {
	ks := make([]string, len(k.keywords))
	i := 0
	for s := range k.keywords {
		ks[i] = s
		i++
	}
	return ks
}
