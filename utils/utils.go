package utils

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
)

// header  \s \t \n

var reBlank *regexp.Regexp
var fws = "(?:(?:\\s*\r?\n)?\\s*)"

func init() {
	reBlank, _ = regexp.Compile("[\\s]+")
}

func Strip(s string) string {
	return strings.TrimFunc(s, isBlank)
}

func StripRight(s string) string {
	return strings.TrimRightFunc(s, isBlank)
}

func StripLeft(s string) string {
	return strings.TrimLeftFunc(s, isBlank)
}

func isBlank(r rune) bool {
	return reBlank.Match([]byte{byte(r)})
}

func RandString(l int) string {
	rand.Seed(time.Now().UnixNano())
	data := make([]byte, l)
	var num int
	for i := 0; i < l; i++ {
		num = rand.Intn(75) + 48
		for {
			if (num > 57 && num < 65) || (num > 90 && num < 97) {
				num = rand.Intn(75) + 48
			} else {
				break
			}
		}
		data[i] = byte(num)
	}
	return string(data)
}
