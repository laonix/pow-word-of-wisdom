package pow

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/itchyny/timefmt-go"
)

const (
	Version      = 1
	FormatDate   = "%y%m%d%H%M"
	FormatHeader = "%d:%d:%s:%s::%s:%s"
)

// Header holds attributes of a Hashcash PoW challenge header.
type Header struct {
	version  uint8 // must be 1
	bits     uint
	date     string // see FormatDate
	resource string
	random   string // base-64 encoded sequence of 10 random bytes
	counter  int64
}

// NewHeader returns a new instance of Header.
func NewHeader(bits uint, resource string) (*Header, error) {
	date := timefmt.Format(time.Now(), FormatDate)

	random, err := getRandom()
	if err != nil {
		return nil, fmt.Errorf("get random: %w", err)
	}

	return &Header{
		version:  Version,
		bits:     bits,
		date:     date,
		resource: resource,
		random:   random,
		counter:  rand.Int63(),
	}, nil
}

// String returns a string representation of Header.
//
// String format must be "%d:%d:%s:%s::%s:%s" (see FormatHeader).
func (h *Header) String() string {
	counter := base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(h.counter, 10)))
	return fmt.Sprintf(FormatHeader, Version, h.bits, h.date, h.resource, h.random, counter)
}

// ParseHeaderString checks an argument header string and returns an instance of Header based on it.
func ParseHeaderString(header string) (*Header, error) {
	split := strings.Split(header, ":")
	if len(split) != 7 {
		return nil, fmt.Errorf("malformed header string [%s]", header)
	}

	version, err := strconv.Atoi(split[0])
	if err != nil {
		return nil, fmt.Errorf("convert version to int: %w", err)
	}
	if version != Version {
		return nil, fmt.Errorf("unsupported version %d", version)
	}

	bits, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, fmt.Errorf("convert bits to int: %w", err)
	}

	date := split[2]
	if _, err := timefmt.Parse(date, FormatDate); err != nil {
		return nil, fmt.Errorf("parse date: %w", err)
	}

	resource := split[3]

	// split[4] stands for omitted extensions

	random := split[5]

	b, err := base64.StdEncoding.DecodeString(split[6])
	if err != nil {
		return nil, fmt.Errorf("decode counter: %w", err)
	}
	counter, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("convert counter to int: %w", err)
	}

	return &Header{
		version:  uint8(version),
		bits:     uint(bits),
		date:     date,
		resource: resource,
		random:   random,
		counter:  counter,
	}, nil
}

// ChallengeFunc is a type of function to generate a Hashcash PoW challenge header string.
type ChallengeFunc func(bits uint, resource string) (string, error)

// Challenge generates a Hashcash PoW challenge header string.
func Challenge(bits uint, resource string) (string, error) {
	header, err := NewHeader(bits, resource)
	if err != nil {
		return "", fmt.Errorf("create new header: %w", err)
	}

	return header.String(), nil
}

// CalculateFunc is a type of function to calculate a Hashcash PoW result header string.
type CalculateFunc func(headerStr string) (string, error)

// we use SHA-256 instead of SHA-1 to keep calculations secured and avoid collisions
var hasher = sha256.New()

// Calculate returns PoW result header string.
//
// The result must have the number of zero leading bits declared in challenge header 'bits' field.
// E.g. if the challenge header is "1:20:2201010000:resource::cmFuZG9t:MTAwMA=="
// than the result must have 20 leading 0 bits.
func Calculate(headerStr string) (string, error) {
	header, err := ParseHeaderString(headerStr)
	if err != nil {
		return "", fmt.Errorf("parse header string: %w", err)
	}

	bits := header.bits
	for {
		calculatedHash := getHash(header.String(), hasher)
		if !checkBits(calculatedHash, bits) {
			header.counter++
			continue
		} else {
			return header.String(), nil
		}
	}
}

// VerifyFunc is a type of function to verify a Hashcash PoW result header string.
type VerifyFunc func(calculated, challenge string) (bool, error)

// Verify checks if the result of PoW calculation is valid:
// it must have the number of zero leading bits declared in challenge header 'bits' field,
// and it must correspond to the challenge header
// (e.g. the difference with the challenge must be in counter field only).
func Verify(calculated, challenge string) (bool, error) {
	calculatedHeader, err := ParseHeaderString(calculated)
	if err != nil {
		return false, fmt.Errorf("parse calculated header string: %w", err)
	}

	challengeHeader, err := ParseHeaderString(challenge)
	if err != nil {
		return false, fmt.Errorf("parse challenge header string: %w", err)
	}

	// check if the calculated PoW result corresponds to the challenge
	if calculatedHeader.version != challengeHeader.version ||
		calculatedHeader.bits != challengeHeader.bits ||
		calculatedHeader.date != challengeHeader.date ||
		calculatedHeader.resource != challengeHeader.resource ||
		calculatedHeader.random != challengeHeader.random {
		return false, errors.New("calculated header doesn't match the challenge")
	}

	// check the number of leading zero bits
	calculatedHash := getHash(calculatedHeader.String(), hasher)
	if !checkBits(calculatedHash, calculatedHeader.bits) {
		return false, nil
	}

	return true, nil
}

func getRandom() (string, error) {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func getHash(header string, hasher hash.Hash) []byte {
	hasher.Reset()
	hasher.Write([]byte(header))
	return hasher.Sum(nil)
}

func checkBits(hash []byte, bits uint) bool {
	modulo := bits % 8
	quotient := bits / 8

	for _, b := range hash[:quotient] {
		if b != 0 {
			return false
		}
	}

	if modulo > 0 {
		b := hash[quotient]
		switch modulo {
		case 1:
			if b > 127 {
				return false
			}
			break
		case 2:
			if b > 63 {
				return false
			}
			break
		case 3:
			if b > 31 {
				return false
			}
			break
		case 4:
			if b > 15 {
				return false
			}
			break
		case 5:
			if b > 7 {
				return false
			}
			break
		case 6:
			if b > 3 {
				return false
			}
			break
		case 7:
			if b > 1 {
				return false
			}
			break
		}
	}

	return true
}
