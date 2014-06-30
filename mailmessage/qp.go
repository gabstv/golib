// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The file define a quoted-printable decoder, as specified in RFC 2045.
// Deviations:
// 1. in addition to "=\r\n", "=\n" is also treated as soft line break.
// 2. it will pass through a '\r' or '\n' not preceded by '=', consistent
//    with other broken QP encoders & decoders.

package mailmessage

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gabstv/latinx"
	"io"
	"strings"
)

type qpReader struct {
	br         *bufio.Reader
	rerr       error  // last read error
	line       []byte // to be consumed before more of br
	isocharset int
}

func (q *qpReader) setcharset(cset string) {
	switch strings.ToLower(cset) {
	case "iso-8859-1":
		q.isocharset = latinx.ISO_8859_1
	}
}

func newQuotedPrintableReaderByCharsetStr(charset string) *qpReader {
	v := &qpReader{
		isocharset: -1,
	}
	v.setcharset(charset)
	return v
}

func newQuotedPrintableReaderByCharsetId(charset int) *qpReader {
	v := &qpReader{
		isocharset: charset,
	}
	return v
}

func fromHex(b byte) (byte, error) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', nil
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, nil
	}
	return 0, fmt.Errorf("multipart: invalid quoted-printable hex byte 0x%02x", b)
}

func (q *qpReader) readHexByte(v []byte) (b byte, err error) {
	if len(v) < 2 {
		return 0, io.ErrUnexpectedEOF
	}
	var hb, lb byte
	if hb, err = fromHex(v[0]); err != nil {
		return 0, err
	}
	if lb, err = fromHex(v[1]); err != nil {
		return 0, err
	}
	return hb<<4 | lb, nil
}

func isQPDiscardWhitespace(r rune) bool {
	switch r {
	case '\n', '\r', ' ', '\t':
		return true
	}
	return false
}

var (
	crlf       = []byte("\r\n")
	lf         = []byte("\n")
	softSuffix = []byte("=")
)

func (q *qpReader) Decode(reader io.Reader, writer io.Writer) (n int, err error) {
	//buffer := new(bytes.Buffer)
	//var n0 int
	rdr := bufio.NewReader(reader)
	for {
		var line []byte
		line, err = rdr.ReadSlice('\n')
		if err != nil {
			if err != io.EOF {
				return
			}
			if len(line) < 1 {
				break
			}
		}
		n += len(line)
		hasLF := bytes.HasSuffix(line, lf)
		hasCR := bytes.HasSuffix(line, crlf)
		wholeLine := line
		line = bytes.TrimRightFunc(wholeLine, isQPDiscardWhitespace)
		if bytes.HasSuffix(line, softSuffix) {
			rightStripped := wholeLine[len(line):]
			line = line[:len(line)-1]
			if !bytes.HasPrefix(rightStripped, lf) && !bytes.HasPrefix((rightStripped), crlf) {
				err = fmt.Errorf("multipart: invalid bytes after =: %q", rightStripped)
				//return ?
			}
		} else if hasLF {
			if hasCR {
				line = append(line, '\r', '\n')
			} else {
				line = append(line, '\n')
			}
		}
		for len(line) > 0 {
			switch {
			case line[0] == '=':
				var b byte
				b, err = q.readHexByte(line[1:])
				if err != nil {
					return
				}
				if q.isocharset > -1 {
					bb, _ := latinx.DecodeByte(q.isocharset, b)
					if bb != nil {
						writer.Write(bb)
					}
				} else {
					writer.Write([]byte{b})
				}
				line = line[3:]
			case line[0] == '\t' || line[0] == '\r' || line[0] == '\n':
				writer.Write([]byte{line[0]})
				line = line[1:]
				break
			case line[0] < ' ' || line[0] > '~':
				err = fmt.Errorf("multipart: invalid unescaped byte 0x%02x in quoted-printable body", line[0])
				return
			default:
				writer.Write([]byte{line[0]})
				line = line[1:]
			}
		}
	}
	err = nil
	return
}
