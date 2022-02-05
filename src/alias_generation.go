package main

import (
	"math/rand"
)

const POSSIBLE_LETTERS = "ABCDEFG"

func GenerateAlias(length int) string {
	res := ""
	for i := 0; i < length; i++ {
		res += string(POSSIBLE_LETTERS[rand.Intn(len(POSSIBLE_LETTERS))])
	}
	return res
}
