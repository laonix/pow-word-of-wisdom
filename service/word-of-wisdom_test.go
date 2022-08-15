package service

//go:generate mockery --name=Getter --case underscore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"github.com/laonix/pow-word-of-wisdom/service/mocks"
)

func TestWordOfWisdomService_Quote(t *testing.T) {
	quotesSource := map[string]string{
		"id_1": "quote_1",
		"id_2": "quote_2",
		"id_3": "quote_3",
	}

	getter := mocks.NewGetter(t)

	for _, id := range maps.Keys(quotesSource) {
		getter.On("Get", id).Maybe().Return(quotesSource[id])
	}

	getter.On("GetIds").Return(maps.Keys(quotesSource))

	srv := NewWordOfWisdomService(getter)

	quote, err := srv.Quote()
	assert.Nil(t, err)
	assert.NotEmpty(t, quote)

	assert.Condition(t, func() (success bool) {
		for _, v := range quotesSource {
			if v == quote {
				return true
			}
		}
		return false
	})

}
