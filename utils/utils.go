package utils

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
)

// header  \s \t \n

var reBlank, reMail *regexp.Regexp
var (
	fwsPattern  = "(?:(?:\\s*\r?\n)?\\s*)"
	mailPattern = "^<([\\w-]+(?:\\.[\\w-]+)*@(?:[\\w-]+\\.)+[a-zA-Z]+)>.*$"
)

func init() {
	reBlank, _ = regexp.Compile("[\\s]+")
	reMail, _ = regexp.Compile(mailPattern)
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

func CutMail(s string) (string, bool) {
	mail := ""
	match := reMail.MatchString(s)
	if match {
		mail = reMail.ReplaceAllString(s, "$1")
	}
	return mail, match
}
