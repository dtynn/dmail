package dkim

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"strconv"
	"strings"
	"time"

	"github.com/dtynn/dmail/message"
)

var bBlank = ""

var bytesColon = []byte(":")
var bytesLineSep = []byte("\r\n")
var DkimHeaderName = "DKIM-Signature"
var bytesLowerDkimHeaderName = []byte(strings.ToLower(DkimHeaderName))

type dkim struct {
	v, // version
	a, // algorithm "rsa-sha256"
	bh, // body hash
	d, // domain "example.com"
	i, // identity "@<domain>"
	q, // query methods "dns/txt"
	s string // selector

	setLength bool
	c         *canon // canonicalization "relaxed/simple"

	l, // body length
	t, // unix timestamp
	x int64 // expiration

	included *dkimHeaders
	musts    map[string]bool
	message  *message.Message
	pemBytes []byte
}

func NewDefaultDkim(message *message.Message, conf *DkimConf) *dkim {
	c := canon{"relaxed", "simple"}
	musts := map[string]bool{}
	for _, must := range headersMust {
		musts[must] = false
	}
	return &dkim{
		v:         "1",
		a:         "rsa-sha256",
		d:         conf.domain,
		i:         conf.identity,
		q:         "dns/txt",
		s:         conf.selector,
		setLength: conf.setLength,
		c:         &c,
		t:         time.Now().Unix(),

		included: newDkimHeader(),
		musts:    musts,
		message:  message,
		pemBytes: conf.pemBytes,
	}
}

func (this *dkim) includeHeaders() {
	for _, mh := range this.message.Headers() {
		field := strings.ToLower(mh.Field())
		if isMust := stringIn(field, headersMust); isMust || stringIn(field, headersShould) {
			// should sign
			if isMust {
				this.musts[field] = true
			}
			canonicalizer, _ := canonHeaders[this.c.header]
			this.included.append(message.NewNormalHeader(mh.Field(), canonicalizer(mh.Value())))
		}
	}
	return
}

func (this *dkim) hashHeaders() hash.Hash {
	toHash := this.included.toHash()
	hasher := sha256.New()
	for _, h := range toHash {
		hasher.Write([]byte(h.Field()))
		hasher.Write(bytesColon)
		hasher.Write([]byte(h.Value()))
		hasher.Write(bytesLineSep)
	}

	hasher.Write(bytesLowerDkimHeaderName)
	hasher.Write(bytesColon)
	hasher.Write([]byte(this.signature(bBlank)))
	return hasher
}

func (this *dkim) hashBody() {
	canonicalizer, _ := canonBody[this.c.body]
	canonedBody := canonicalizer(string(this.message.EncodeBody()))
	if this.setLength {
		this.l = int64(len(canonedBody))
	}
	this.bh = generateSha256Hash(canonedBody)
}

func (this *dkim) tags(bv string) []string {
	tags := []string{
		makeDkimTag("v", this.v),
		makeDkimTag("a", this.a),
		makeDkimTag("c", this.c.String()),
		makeDkimTag("d", this.d),
		makeDkimTag("i", fmt.Sprintf("%s@%s", this.i, this.d)),
		makeDkimTag("q", this.q),
		makeDkimTag("s", this.s),
		makeDkimTag("t", strconv.Itoa(int(this.t))),
	}
	if this.setLength {
		tags = append(tags, makeDkimTag("l", strconv.Itoa(int(this.l))))
	}
	tags = append(tags,
		makeDkimTag("h", this.included.h()),
		makeDkimTag("bh", this.bh),
		makeDkimTag("b", bv))
	return tags
}

func (this *dkim) signature(bv string) string {
	return strings.Join(this.tags(bv), "; ")
}

func (this *dkim) checkMusts() bool {
	for _, ok := range this.musts {
		if !ok {
			return false
		}
	}
	return true
}

func (this *dkim) Sign() (string, error) {
	// get body and hash
	this.hashBody()
	// get headers
	this.includeHeaders()
	// hash headers
	if checked := this.checkMusts(); !checked {
		return "", fmt.Errorf("not all MUST headers included")
	}

	signer, err := newSigner(this.pemBytes)
	if err != nil {
		return "", err
	}
	bv, err := signer.SignHash(this.hashHeaders())
	if err != nil {
		return "", err
	}
	return foldHeader(this.signature(bv)), nil
}

func (this *dkim) SignatureHeader() (message.Header, error) {
	sig, err := this.Sign()
	if err != nil {
		return nil, err
	}

	return message.NewNormalHeader(DkimHeaderName, sig), nil
}
