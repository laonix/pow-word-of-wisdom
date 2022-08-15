package service

import (
	_ "embed"
	"encoding/json"
	"errors"
	"math/rand"
	"sync"

	"golang.org/x/exp/maps"
)

// WordOfWisdom is a contract to get a word of wisdom quote.
type WordOfWisdom interface {
	Quote() (string, error)
}

// WordOfWisdomService is an implementation of WordOfWisdom.
//
// It returns a random quote from a quotes source.
type WordOfWisdomService struct {
	getter Getter
	ids    *IdsHolder
}

// NewWordOfWisdomService returns a new instance of WordOfWisdomService.
func NewWordOfWisdomService(getter Getter) *WordOfWisdomService {
	return &WordOfWisdomService{
		getter: getter,
		ids:    &IdsHolder{ids: getter.GetIds()},
	}
}

// Quote returns a random word of wisdom quote.
func (src *WordOfWisdomService) Quote() (string, error) {
	n := rand.Intn(src.ids.Len())
	id, ok := src.ids.Get(n)
	if !ok {
		return "", errors.New("get random quote id")
	}

	return src.getter.Get(id), nil
}

// Getter is a contract to get a quote from some source.
type Getter interface {
	Get(id string) string
	GetIds() []string
}

// FileGetter is an implementation of Getter to retrieve quotes from file.
type FileGetter struct {
	rw     sync.RWMutex
	quotes map[string]string
}

//go:embed recource/quote.json
var quoteBytes []byte

// NewFileGetter returns a new instance of FileGetter.
func NewFileGetter() *FileGetter {
	var tmp map[string]string
	if err := json.Unmarshal(quoteBytes, &tmp); err != nil {
		return &FileGetter{quotes: make(map[string]string, 0)}
	}

	return &FileGetter{quotes: tmp}
}

// Get returns a quote string by its id.
func (g *FileGetter) Get(id string) string {
	g.rw.RLock()
	defer g.rw.RUnlock()

	return g.quotes[id]
}

// GetIds returns a set of stored quotes ids.
func (g *FileGetter) GetIds() []string {
	g.rw.RLock()
	defer g.rw.RUnlock()

	return maps.Keys(g.quotes)
}

// IdsHolder holds a set of quotes ids.
type IdsHolder struct {
	rw  sync.RWMutex
	ids []string
}

// Get returns a quote id by its id in IdsHolder.
func (ih *IdsHolder) Get(idKey int) (string, bool) {
	ih.rw.RLock()
	defer ih.rw.RUnlock()

	if idKey < 0 || idKey >= len(ih.ids) {
		return "", false
	}

	return ih.ids[idKey], true
}

// Len returns a count of held quotes (quotes' ids).
func (ih *IdsHolder) Len() int {
	return len(ih.ids)
}
