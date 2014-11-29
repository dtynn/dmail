package dkim

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"regexp"
	"strings"
)

// header  \s \t \n

var Hasher = map[string]func(string) string{"sha1": generateSha1Hash, "sha256": generateSha256Hash}

var reCRLF, reMultiWS, reTrailingLines, reTrailingWS, reBlank, reBTag *regexp.Regexp
var FWS = "(?:(?:\\s*\r?\n)?\\s*)"

func init() {
	reCRLF, _ = regexp.Compile("\r\n")
	reMultiWS, _ = regexp.Compile("[\t ]+")
	reTrailingLines, _ = regexp.Compile("(\r\n)*$")
	reTrailingWS, _ = regexp.Compile("[\t ]+\r\n")
	reBlank, _ = regexp.Compile("[\\s]+")
	reBTag, _ = regexp.Compile("([;\\s]b" + FWS + "?=)(?:" + FWS + "[a-zA-Z0-9+/=])*(?:\r?\n$)?")
}

func foldHeader(s string) string {
	pre := ""
	i := strings.LastIndex(s, "\r\n ")
	if i != -1 {
		i += 3
		pre = s[:i]
		s = s[i:]
	}

	for len(s) > 72 {
		j := 72
		if i := strings.LastIndex(s[:72], " "); i != -1 {
			j = i + 1
		}
		pre += s[:j] + "\r\n "
		s = s[j:]
	}
	return pre + s
}

func unfoldHeader(src string) string {
	return reCRLF.ReplaceAllString(src, "")
}

func stripTrailingLine(src string) string {
	return reTrailingLines.ReplaceAllString(src, "\r\n")
}

func compressWhitespace(src string) string {
	return reMultiWS.ReplaceAllString(src, " ")
}

func stripTrailingWhitespace(src string) string {
	return reTrailingWS.ReplaceAllString(src, "\r\n")
}

func generateSha1Hash(s string) string {
	return generateHash(s, sha1.New())
}

func generateSha256Hash(s string) string {
	return generateHash(s, sha256.New())
}

func generateHash(s string, h hash.Hash) string {
	h.Write([]byte(s))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func stringIn(val string, slice []string) bool {
	if len(slice) == 0 {
		return false
	}
	for _, one := range slice {
		if val == one {
			return true
		}
	}
	return false
}

func makeDkimTag(name string, value string) string {
	return fmt.Sprintf("%s=%s", name, value)
}

func cutBTag(src string) string {
	return reBTag.ReplaceAllString(src, "$1")
}
