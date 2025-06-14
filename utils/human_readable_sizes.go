package utils

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"unicode"
)

var (
	unitsSlice = []byte("kmgtp")
)

const (
	toLowerTable = "\x00\x01\x02\x03\x04\x05\x06\a\b\t\n\v\f\r\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\"#$%&'()*+,-./0123456789:;<=>?@abcdefghijklmnopqrstuvwxyz[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~\u007f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8\xc9\xca\xcb\xcc\xcd\xce\xcf\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf\xe0\xe1\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"
	toUpperTable = "\x00\x01\x02\x03\x04\x05\x06\a\b\t\n\v\f\r\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`ABCDEFGHIJKLMNOPQRSTUVWXYZ{|}~\u007f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8\xc9\xca\xcb\xcc\xcd\xce\xcf\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf\xe0\xe1\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"
)

// HumanReadableSizeToBytes returns integer size of bytes from human-readable string, ex. 42kb, 42M
// Returns 0 if string is unrecognized
func HumanReadableSizeToBytes(humanReadableString string) int {
	strLen := len(humanReadableString)
	if strLen == 0 {
		return 0
	}
	var unitPrefixPos, lastNumberPos int
	// loop the string
	for i := strLen - 1; i >= 0; i-- {
		// check if the char is a number
		if unicode.IsDigit(rune(humanReadableString[i])) {
			lastNumberPos = i
			break
		} else if humanReadableString[i] != ' ' {
			unitPrefixPos = i
		}
	}

	// fetch the number part and parse it to float
	size, err := strconv.ParseFloat(humanReadableString[:lastNumberPos+1], 64)
	if err != nil {
		return 0
	}

	// check the multiplier from the string and use it
	if unitPrefixPos > 0 {
		// convert multiplier char to lowercase and check if exists in units slice
		index := bytes.IndexByte(unitsSlice, toLowerTable[humanReadableString[unitPrefixPos]])
		if index != -1 {
			size *= math.Pow(1000, float64(index+1))
		}
	}

	return int(size)
}

var units = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

// ConvertBytesToHumanReadableSize converts the given integer size in bytes to a human-readable string format.
// If the size is 0, it returns "0B".
// It uses the logarithmic calculation to determine the appropriate unit prefix (e.g., KB, MB, GB, etc.),
// and formats the value with two decimal places before appending the unit suffix.
// The units slice contains the available unit prefixes: ["B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"].
// Note: The function assumes that the given size is a positive integer.
func ConvertBytesToHumanReadableSize(bytes int) string {
	if bytes == 0 {
		return "0B"
	}

	i := int(math.Floor(math.Log(float64(bytes)) / math.Log(1000)))
	suffix := units[i]
	value := float64(bytes) / math.Pow(1000, float64(i))

	return fmt.Sprintf("%.2f%s", value, suffix)
}
