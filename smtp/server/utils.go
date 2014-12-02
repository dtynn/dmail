package server

import (
	"regexp"
)

var reMail *regexp.Regexp
var mailPattern = "^<[\\w-]+(\\.[\\w-]+)*@([\\w-]+\\.)+[a-zA-Z]+>"

func init() {
	reMail, _ = regexp.Compile(mailPattern)
}
