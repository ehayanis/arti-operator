package utils

import (
	"math/rand"
	"strings"
	"time"
)

func GenerateRandomPassword(length int) string {
	rand.Seed(time.Now().UnixNano())

	// Removed usual ambiguous chars : one/lowercase_l, zero/letter_O
	chars := []rune("ABCDEFGHIJKLMNPQRSTUVWXYZ" +
		"abcdefghijkmnpqrstuvwxyz" +
		"23456789" +
		"/_-!")

	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()

	return str
}
