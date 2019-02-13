package utils

import (
	"math/rand"
	"strings"
	"time"
)

func GenerateRandomPassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNPQRSTUVWXYZ" +
		"abcdefghijkmnpqrstuvwxyz" +
		"23456789" +
		"/_-!")

	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String() // E.g. "ExcbsVQs"

	return str
}
