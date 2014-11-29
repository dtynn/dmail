package dkim

import (
	"strings"

	"github.com/dtynn/dmail/message"
)

var headersMust = []string{
	"from",
}

var headersShould = []string{
	"sender",
	"reply-to",
	"subject",
	"date",
	"message-id",
	"to",
	"cc",
	"mime-version",
	"content-type",
	"content-transfer-encoding",
	"content-id",
	"content- description",
	"resent-date",
	"resent-from",
	"resent-sender",
	"resent-to",
	"resent-cc",
	"resent-message-id",
	"in-reply-to",
	"references",
	"list-id",
	"list-help",
	"list-unsubscribe",
	"list-subscribe",
	"list-post",
	"list-owner",
	"list-archive",
}

var headersShouldNot = []string{
	"return-path",
	"received",
	"comments",
	"keywords",
	"bcc",
	"resent-bcc",
	"dkim-signature",
}

var headersFrozen = []string{
	"from",
	"subject",
	"date",
}

type dkimHeaders struct {
	hs []message.Header
}

func newDkimHeader() *dkimHeaders {
	return &dkimHeaders{[]message.Header{}}
}

func (this *dkimHeaders) append(h message.Header) {
	this.hs = append(this.hs, h)
}

func (this *dkimHeaders) fields() []string {
	f := make([]string, len(this.hs))
	for i, h := range this.hs {
		f[i] = h.Field()
	}
	return f
}

func (this *dkimHeaders) frozen() []string {
	frozened := []string{}
	for _, f := range this.fields() {
		// should frozen but not in frozened
		if stringIn(strings.ToLower(f), headersFrozen) {
			frozened = append(frozened, f)
		}
	}
	return frozened
}

func (this *dkimHeaders) toHash() []message.Header {
	l := len(this.hs)
	headersToHash := make([]message.Header, l)
	lastIndex := map[string]int{}

	for i, h := range this.hs {
		field := strings.ToLower(h.Field())
		last, ok := lastIndex[field]
		if !ok {
			last = l
		}
		for last > 0 {
			last -= 1
			if field == strings.ToLower(this.hs[last].Field()) {
				lastIndex[field] = last
				headersToHash[i] = message.NewNormalHeader(field, this.hs[last].Value())
				break
			}
		}
	}

	return headersToHash
}

func (this *dkimHeaders) h() string {
	all := append(this.fields(), this.frozen()...)
	return strings.Join(all, " : ")
}
