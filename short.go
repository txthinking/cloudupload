package cloudupload

import (
	"math"
	"strings"
)

const (
	SYMBOLS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	BASE    = int64(len(SYMBOLS))
)

// encodes a number into our *base* representation
func ToShort(number int64) string {
	rest := number % BASE
	// strings are a bit weird in go...
	result := string(SYMBOLS[rest])
	if number-rest != 0 {
		newnumber := (number - rest) / BASE
		result = ToShort(newnumber) + result
	}
	return result
}

// Decodes a string given in our encoding and returns the decimal
// integer.
func FromShort(input string) int64 {
	const floatbase = float64(BASE)
	l := len(input)
	var sum int = 0
	for index := l - 1; index > -1; index -= 1 {
		current := string(input[index])
		pos := strings.Index(SYMBOLS, current)
		sum = sum + (pos * int(math.Pow(floatbase, float64((l-index-1)))))
	}
	return int64(sum)
}
