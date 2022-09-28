package main

import (
	"fmt"
	"regexp"
)

func main() {
	key := "lkjhfsdfsd982734"
	fmt.Printf("result of validating %s, %t", key, ValidateUUID(key))
}

func ValidateUUID(key string) bool {
	reg := regexp.MustCompile("^[A-Za-z\\d_-]*$")

	return reg.MatchString(key)
}
