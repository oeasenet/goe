// Identifier generators: NanoID, UUIDv7 and XID.

package utils

import (
	"crypto/rand"
	"errors"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/xid"
)

// GenerateNanoId returns a cryptographically random, URL-safe identifier built
// from an alphanumeric alphabet. The default length is 32 characters; pass a
// single argument to choose another.
//
// It panics if the underlying random source fails, which in practice means the
// operating system's entropy pool is unavailable.
func GenerateNanoId(length ...int) string {
	// defaultAlphabet is the alphabet used for ID characters by default.
	const defaultAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	size := 32
	if len(length) > 0 {
		size = length[0]
	}
	// getMask generates bit mask used to obtain bits from the random bytes that are used to get index of random character
	// from the alphabet. Example: if the alphabet has 6 = (110)_2 characters it is sufficient to use mask 7 = (111)_2
	var getMask = func(alphabetSize int) int {
		for i := 1; i <= 8; i++ {
			mask := (2 << uint(i)) - 1
			if mask >= alphabetSize-1 {
				return mask
			}
		}
		return 0
	}

	// Generate is a low-level function to change alphabet and ID size.
	var generate = func(alphabet string, size int) (string, error) {
		chars := []rune(alphabet)

		if len(alphabet) == 0 || len(alphabet) > 255 {
			return "", errors.New("alphabet must not be empty and contain no more than 255 chars")
		}
		if size <= 0 {
			return "", errors.New("size must be positive integer")
		}

		mask := getMask(len(chars))
		// to estimate how many random bytes we will need for the ID, we might actually need more but this is tradeoff
		// between an average case and the worst case
		ceilArg := 1.6 * float64(mask*size) / float64(len(alphabet))
		step := int(math.Ceil(ceilArg))

		id := make([]rune, size)
		bytes := make([]byte, step)
		for j := 0; ; {
			_, err := rand.Read(bytes)
			if err != nil {
				return "", err
			}
			for i := range step {
				currByte := bytes[i] & byte(mask)
				if currByte < byte(len(chars)) {
					id[j] = chars[currByte]
					j++
					if j == size {
						return string(id[:size]), nil
					}
				}
			}
		}
	}

	result, err := generate(defaultAlphabet, size)
	if err != nil {
		panic(err)
	}
	return result
}

// GenerateUUIDv7 returns a UUID version 7 with the dashes stripped, or an empty
// string if one could not be generated. Version 7 is time-ordered, so the values
// sort chronologically — useful as a database primary key.
func GenerateUUIDv7() string {
	uuidWithDashes, err := uuid.NewV7()
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(uuidWithDashes.String(), "-", "")
}

// GenerateXid returns a globally unique, 20-character, time-sortable identifier.
// XIDs are shorter than UUIDs and encode a timestamp, machine and process id.
func GenerateXid() string {
	return xid.New().String()
}
