package dmail

import (
	"bytes"
	"io"
)

// As defined in RFC 5322, 2.1.1.
const maxLineLen = 78

// base64LineWriter limits text encoded in base64 to 78 characters per line
type base64LineWriter struct {
	w       io.Writer
	lineLen int
}

func newBase64LineWriter(w io.Writer) *base64LineWriter {
	return &base64LineWriter{w: w}
}

func (w *base64LineWriter) Write(p []byte) (int, error) {
	n := 0
	for len(p)+w.lineLen > maxLineLen {
		w.w.Write(p[:maxLineLen-w.lineLen])
		w.w.Write([]byte("\r\n"))
		p = p[maxLineLen-w.lineLen:]
		n += maxLineLen - w.lineLen
		w.lineLen = 0
	}

	w.w.Write(p)
	w.lineLen += len(p)

	return n + len(p), nil
}

// qpLineWriter limits text encoded in quoted-printable to 78 characters per
// line
type qpLineWriter struct {
	w       io.Writer
	lineLen int
}

func newQpLineWriter(w io.Writer) *qpLineWriter {
	return &qpLineWriter{w: w}
}

func (w *qpLineWriter) Write(p []byte) (int, error) {
	n := 0
	for len(p) > 0 {
		// If the text is not over the limit, write everything
		if len(p) < maxLineLen-w.lineLen {
			w.w.Write(p)
			w.lineLen += len(p)
			return n + len(p), nil
		}

		i := bytes.IndexAny(p[:maxLineLen-w.lineLen+2], "\n")
		// If there is a newline before the limit, write the end of the line
		if i != -1 && (i != maxLineLen-w.lineLen+1 || p[i-1] == '\r') {
			w.w.Write(p[:i+1])
			p = p[i+1:]
			n += i + 1
			w.lineLen = 0
			continue
		}

		// Quoted-printable text must not be cut between an equal sign and the
		// two following characters
		var toWrite int
		if maxLineLen-w.lineLen-2 >= 0 && p[maxLineLen-w.lineLen-2] == '=' {
			toWrite = maxLineLen - w.lineLen - 2
		} else if p[maxLineLen-w.lineLen-1] == '=' {
			toWrite = maxLineLen - w.lineLen - 1
		} else {
			toWrite = maxLineLen - w.lineLen
		}

		// Insert the newline where it is needed
		w.w.Write(p[:toWrite])
		w.w.Write([]byte("=\r\n"))
		p = p[toWrite:]
		n += toWrite
		w.lineLen = 0
	}

	return n, nil
}
