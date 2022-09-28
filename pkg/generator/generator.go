package generator

import (
	"math"

	"github.com/chr4/pwgen"
)

func GenerateSecret(len int64) string {
	if len < math.MaxInt {
		str := pwgen.AlphaNumSymbols(int(len))

		return str
	}
	return ""
}

func GenerateKey(len int64) string {
	if len < math.MaxInt {
		str := pwgen.AlphaNum(int(len))

		return str
	}
	return ""
}
