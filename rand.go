package goutil

import (
	"math/rand"
	"time"
)

const (
	lower = "abcdefghijklmnopqrstuvyxyz"
	upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digit = "1234567890"
)

func GetRandom(n int, up, lo, digi bool) string {
	var char string
	if !up && !lo && !digi {
		char = lower + upper + digit
	}
	if up {
		char += upper
	}
	if lo {
		char += lower
	}
	if digi {
		char += digit
	}
	set := []rune(char)
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = set[rand.Intn(len(set))]
	}
	return string(b)
}
