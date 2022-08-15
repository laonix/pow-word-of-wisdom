package pow

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/itchyny/timefmt-go"
	"github.com/stretchr/testify/assert"
)

func TestChallenge(t *testing.T) {
	assertions := assert.New(t)

	bits := 5
	resource := "resource"

	header, err := Challenge(uint(bits), resource)
	assertions.Nil(err)
	assertions.NotEmpty(header)

	split := strings.Split(header, ":")
	assertions.Len(split, 7)

	v, err := strconv.Atoi(split[0])
	assertions.Nil(err)
	assertions.EqualValues(v, Version)

	b, err := strconv.Atoi(split[1])
	assertions.Nil(err)
	assertions.EqualValues(b, bits)

	assertions.Equal(split[2], timefmt.Format(time.Now(), FormatDate))

	assertions.Equal(split[3], resource)

	assertions.Empty(split[4])

	assertions.NotEmpty(split[5])

	bytes, err := base64.StdEncoding.DecodeString(split[6])
	assertions.Nil(err)
	_, err = strconv.ParseInt(string(bytes), 10, 64)
	assertions.Nil(err)
}

func TestCalculate_correct(t *testing.T) {
	challenge := "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyNDEzOA=="
	expectedResult := "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyODcyOA=="

	result, err := Calculate(challenge)
	assert.Nil(t, err)
	assert.Equal(t, result, expectedResult)
}

func TestCalculate_error(t *testing.T) {
	challenge := "corrupted"

	result, err := Calculate(challenge)
	assert.NotNil(t, err)
	assert.Regexp(t, "parse header string:*", err.Error())
	assert.Empty(t, result)
}

func TestVerify(t *testing.T) {
	tests := []struct {
		name       string
		challenge  string
		calculated string
		want       bool
		err        error
	}{
		{
			name:       "correct calculation result",
			challenge:  "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyNDEzOA==",
			calculated: "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyODcyOA==",
			want:       true,
			err:        nil,
		},
		{
			name:       "incorrect calculation result",
			challenge:  "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyNDEzOA==",
			calculated: "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyODcyNw==",
			want:       false,
			err:        nil,
		},
		{
			name:       "calculated result doesn't match the challenge",
			challenge:  "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyNDEzOA==",
			calculated: "1:12:2208082127:f1a5a003-27ce-4e62-8c48-14c250965b92::kUumfNZAqta03Q==:MTA4MDAyODM5MTgzMzgyMTg0OQ==",
			want:       false,
			err:        errors.New("calculated header doesn't match the challenge"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Verify(test.calculated, test.challenge)
			assert.Equal(t, err, test.err)
			assert.Equal(t, got, test.want)
		})
	}
}

func TestParseHeaderString_correct(t *testing.T) {
	assertions := assert.New(t)

	header, err := ParseHeaderString("1:2:2201010000:resource::cmFuZG9t:MTAwMA==")
	assertions.Nil(err)

	if assertions.NotNil(header) {
		assertions.EqualValues(header.version, Version)
		assertions.EqualValues(header.bits, 2)
		assertions.Equal(header.date, "2201010000")
		assertions.Equal(header.resource, "resource")
		assertions.Equal(header.random, "cmFuZG9t")
		assertions.EqualValues(header.counter, 1000)
	}
}

func TestParseHeaderString_error(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		errRegexp string
	}{
		{
			name:      "malformed header",
			header:    "corrupted",
			errRegexp: "malformed header string*",
		},
		{
			name:      "incorrect version",
			header:    "duck:2:2201010000:resource::cmFuZG9t:MTAwMA==",
			errRegexp: "convert version to int*",
		},
		{
			name:      "unsupported version",
			header:    "3:2:2201010000:resource::cmFuZG9t:MTAwMA==",
			errRegexp: "unsupported version*",
		},
		{
			name:      "incorrect bits",
			header:    "1:duck:2201010000:resource::cmFuZG9t:MTAwMA==",
			errRegexp: "convert bits to int*",
		},
		{
			name:      "incorrect date format",
			header:    "1:2:2022-01-01T00-00:resource::cmFuZG9t:MTAwMA==",
			errRegexp: "parse date*",
		},
		{
			name:      "undecodable counter",
			header:    "1:2:2201010000:resource::cmFuZG9t:*#$*",
			errRegexp: "decode counter*",
		},
		{
			name:      "incorrect counter",
			header:    "1:2:2201010000:resource::cmFuZG9t:duck",
			errRegexp: "convert counter to int*",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			header, err := ParseHeaderString(test.header)
			assert.NotNil(t, err)
			assert.Regexp(t, test.errRegexp, err.Error())
			assert.Nil(t, header)
		})
	}
}

func TestHeader_String(t *testing.T) {
	header, err := NewHeader(2, "resource")
	assert.Nil(t, err)
	assert.NotEmpty(t, header)

	assert.Regexp(t, "1:2:"+timefmt.Format(time.Now(), FormatDate)+":resource::*:*", header.String())
}
