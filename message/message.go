package message

import (
	"bytes"
	"encoding/base64"
	"time"

	"github.com/dtynn/dmail/utils"
	"gopkg.in/alexcesaro/quotedprintable.v1"
)

type Charset string
type Encoding string

const (
	QuotedPrintable Encoding = "quoted-printable"
	Base64          Encoding = "base64"
	Unencoded       Encoding = "8bit"
)

const (
	CharsetUTF8 Charset = "UTF-8"
)

type Message struct {
	headers []Header
	body    string

	charset     Charset
	encoding    Encoding
	contentType string

	hEncoder *quotedprintable.HeaderEncoder
}

func NewMessage(encoding Encoding, charset Charset, contentType string) *Message {
	m := Message{
		headers:     make([]Header, 0),
		charset:     charset,
		encoding:    encoding,
		contentType: contentType,
	}

	var e quotedprintable.Encoding
	switch encoding {
	case Base64:
		e = quotedprintable.B
	case QuotedPrintable:
		e = quotedprintable.Q
	}
	m.hEncoder = e.NewHeaderEncoder(string(charset))
	return &m
}

func (this *Message) AddHeader(h Header) {
	this.headers = append(this.headers, h)
}

func (this *Message) AddNormalHeader(field, value string) {
	this.AddHeader(&NormalHeader{utils.Strip(field), this.encodeHeader(value)})
}

func (this *Message) AddAddressHeader(field, address, name string) {
	this.AddHeader(&AddressHeader{utils.Strip(field), address, this.encodeHeader(name)})
}

func (this *Message) AddDateHeader(field string, date time.Time) {
	this.AddHeader(&DateHeader{utils.Strip(field), date})
}

func (this *Message) encodeHeader(value string) string {
	return this.hEncoder.Encode(value)
}

func (this *Message) SetBody(body string) {
	this.body = body
}

func (this *Message) AddContentType() {
	this.AddNormalHeader("Content-Type", this.contentType+"; charset="+string(this.charset))
}

func (this *Message) AddTransferEncodingHeader() {
	this.AddNormalHeader("Content-Transfer-Encoding", string(this.encoding))
}

func (this *Message) AddDate() {
	now := time.Now()
	this.AddDateHeader("Date", now)
}

func (this *Message) Headers() []Header {
	return this.headers
}

func (this *Message) EncodeBody() []byte {
	buf := new(bytes.Buffer)
	p := []byte(this.body)
	switch this.encoding {
	case Base64:
		writer := base64.NewEncoder(base64.StdEncoding, newBase64LineWriter(buf))
		writer.Write(p)
		writer.Close()
	case QuotedPrintable:
		writer := quotedprintable.NewEncoder(newQpLineWriter(buf))
		writer.Write(p)
	default:
		buf.Write(p)
	}
	return buf.Bytes()
}

func (this *Message) Bytes() []byte {
	buf := new(bytes.Buffer)
	for _, h := range this.headers {
		buf.WriteString(h.String() + "\r\n")
	}
	buf.WriteString("\r\n")
	buf.Write(this.EncodeBody())
	buf.WriteString("\r\n")
	return buf.Bytes()
}
