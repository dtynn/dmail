package dmail

var canonHeaders = map[string]func(h string) string{"simple": simpleHeader, "relaxed": relaxedHeader}
var canonBody = map[string]func(body string) string{"simple": simpleBody, "relaxed": relaxedBody}

func simpleHeader(h string) string {
	return h
}

func simpleBody(body string) string {
	return stripTrailingLine(body)
}

func relaxedHeader(h string) string {
	return compressWhitespace(strip(unfoldHeader(h)))
}

func relaxedBody(body string) string {
	return stripTrailingLine(compressWhitespace(stripTrailingWhitespace(body)))
}

type canon struct {
	header, body string
}

func (this *canon) String() string {
	return this.header + "/" + this.body
}
