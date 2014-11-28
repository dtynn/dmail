package dmail

import (
	"bytes"
	"encoding/base64"
	"time"

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
	headers []header
	body    string

	charset  Charset
	encoding Encoding

	hEncoder *quotedprintable.HeaderEncoder
}

func NewMessage(encoding Encoding, charset Charset) *Message {
	m := Message{
		headers:  make([]header, 0),
		charset:  charset,
		encoding: encoding,
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

func (this *Message) AddHeader(h header) {
	this.headers = append(this.headers, h)
}

func (this *Message) AddNormalHeader(field, value string) {
	this.AddHeader(&normalHeader{strip(field), this.encodeHeader(value)})
}

func (this *Message) AddAddressHeader(field, address, name string) {
	this.AddHeader(&addressHeader{strip(field), address, this.encodeHeader(name)})
}

func (this *Message) AddDateHeader(field string, date time.Time) {
	this.AddHeader(&dateHeader{strip(field), date})
}

func (this *Message) encodeHeader(value string) string {
	return this.hEncoder.Encode(value)
}

func (this *Message) SetBody(body string) {
	this.body = body
}

func (this *Message) AddContentType(contentType string) {
	this.AddNormalHeader("Content-Type", contentType+"; charset="+string(this.charset))
}

func (this *Message) AddTransferEncodingHeader() {
	this.AddNormalHeader("Content-Transfer-Encoding", string(this.encoding))
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
