package mailmessage

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/gabstv/golib/mail-header"
	"io"
	"io/ioutil"
	"log"
	"net/mail"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	CRLF               = "\r\n"
	MSG_MULTIPARTMIXED = iota
	MSG_MULTIPARTALTERNATIVE
	MSG_MESSAGE
)

var (
	utf8b      = regexp.MustCompile(`(?i)=\?UTF-8\?B\?(.*?)\?=`)
	utf8q      = regexp.MustCompile(`(?i)=\?UTF-8\?Q\?(.*?)\?=`)
	iso88591q  = regexp.MustCompile(`(?i)\?iso-8859-1\?Q\?(.*?)\?`)
	iso88591q2 = regexp.MustCompile(`(?i)=\?iso-8859-1\?Q\?(.*?)\?=`)
)

type Message struct {
	Kind     int
	Header   mail.Header
	tempbody io.Reader
	Path     string
	File     *os.File
	Children []*Message
	decoded  bool
}

func (m *Message) DebugPrint() {
	log.Println("m *Message DebugPrint [HEADER CT]", m.Header.Get("Content-Type"))
	if len(m.Header.Get("From")) > 0 {
		log.Println("m *Message DebugPrint [HEADER FROM]", m.Header.Get("From"))
	}
	if len(m.Header.Get("Subject")) > 0 {
		log.Println("m *Message DebugPrint [HEADER SUBJECT]", m.Header.Get("Subject"))
	}
	if len(m.Header.Get("X-Original-From")) > 0 {
		log.Println("m *Message DebugPrint [HEADER X-Original-From]", m.Header.Get("X-Original-From"))
	}
	//log.Println("m *Message DebugPrint [HEADER]", m.Header)
	if m.File != nil {
		len0, _ := io.Copy(ioutil.Discard, m.File)
		m.File.Seek(0, 0)
		log.Println("m *Message DebugPrint [BODY]", len0, "bytes")
	}
	if m.Children != nil {
		log.Println("m *Message DebugPrint [CHILDREN START]")
		for _, v := range m.Children {
			v.DebugPrint()
		}
		log.Println("m *Message DebugPrint [CHILDREN STOP]")
	}
}

func (m *Message) AllMessages() []*Message {
	mmsg := make([]*Message, 0)
	mmsg = append(mmsg, m)
	if m.Children != nil {
		for _, v := range m.Children {
			list0 := v.AllMessages()
			mmsg = append(mmsg, list0...)
		}
	}
	return mmsg
}

func (m *Message) Purge() {
	if m.File != nil {
		m.File.Close()
		m.File = nil
	}
	if len(m.Path) > 0 {
		os.Remove(m.Path)
	}
	if m.Children != nil {
		for k := range m.Children {
			m.Children[k].Purge()
		}
	}
}

func findhtml(m *Message) string {
	b := new(bytes.Buffer)
	if m.Children != nil {
		for _, v := range m.Children {
			b.WriteString(findhtml(v))
		}
	}
	if ct := m.Header.Get("Content-Type"); strings.HasPrefix(ct, "text/html") {
		io.Copy(b, m.File)
		m.File.Seek(0, 0)
	}
	return b.String()
}

func findplaintext(m *Message) string {
	b := new(bytes.Buffer)
	if m.Children != nil {
		for _, v := range m.Children {
			b.WriteString(findplaintext(v))
		}
	}
	if ct := m.Header.Get("Content-Type"); strings.HasPrefix(ct, "text/plain") {
		io.Copy(b, m.File)
		m.File.Seek(0, 0)
	}
	return b.String()
}

func (m *Message) HTML() string {
	return findhtml(m)
}

func (m *Message) Plaintext() string {
	return findplaintext(m)
}

func New(rdr *bufio.Reader) (*Message, error) {

	line, err := rdr.ReadString('\n')
	if err != nil {
		return nil, errors.New("ERR: Newline: " + err.Error())
	}

	if strings.HasPrefix(line, "+OK") {
		log.Println("NEW MAIL", line[4:])
	} else if strings.HasPrefix(line, "-ERR") {
		return nil, errors.New("strings.HasPrefix(line, \"-ERR\") " + line[5:])
	} else {
		//TEMP
		//b00 := new(bytes.Buffer)
		//io.Copy(b00, rdr)
		return nil, errors.New("Unknown pop3 server response `" + line + "`") // `" + b00.String() + "`")
	}
	// save to a temporary file
	nowf := "msg_" + strconv.FormatInt(time.Now().Unix(), 10) + ".dat"
	fil0, err := os.OpenFile(path.Join(os.TempDir(), nowf), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)

	if err != nil {
		return nil, errors.New("os.OpenFile(path.Join(os.TempDir(), nowf), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666) " + err.Error())
	}

	for {
		line, err = rdr.ReadString('\n')
		if err != nil {
			return nil, errors.New("line, err = rdr.ReadString('\n') " + err.Error())
		}
		if line == "."+CRLF {
			break
		}
		_, err = fil0.WriteString(line)
		if err != nil {
			log.Println("ERROROROR", err)
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
	}
Rummage:
	fil0.Seek(0, 0)
	mainm, err := mail.ReadMessage(fil0)
	if err != nil {
		erstr := err.Error()
		if strings.HasPrefix(erstr, "malformed MIME header line: ") {
			// remove the line
			//TODO: see if this is the right way to treat it
			fil1p, fil1f, _ := tempFile()
			fil0.Seek(0, 0)
			io.Copy(fil1f, fil0)
			fil1f.Seek(0, 0)
			bio1 := bufio.NewReader(fil1f)
			fil0.Seek(0, 0)
			fil0.Truncate(0)
			fil0.Seek(0, 0)
			for {
				line, lerr := bio1.ReadString('\n')
				if lerr != nil {
					break
				}
				if strings.HasPrefix(line, erstr[28:]) {
					log.Println("FOUND THE CULPRIT HEADER LINE!", line)
				} else {
					fil0.WriteString(line)
				}
			}
			fil0.Seek(0, 0)
			fil1f.Close()
			os.Remove(fil1p)
			goto Rummage
		}
		log.Println("ERROR ON READMESSAGE")
		return nil, errors.New("ERROR ON READMESSAGE: " + err.Error())
	}
	mainm.Header = normalizeHeaders(mainm.Header)
	contentType := mainm.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/mixed") || strings.HasPrefix(contentType, "multipart/related") {
		return multipartMessage(mainm, fil0)
	} else if strings.HasPrefix(contentType, "multipart/alternative") {
		return alternativeMessage(mainm, fil0)
	}
	return basicMessage(mainm, fil0)
}

// Remove all temporary files related to it.
func (m *Message) Destroy() {
	if m.Children != nil {
		for _, v := range m.Children {
			v.Destroy()
		}
	}
	if m.File != nil {
		n := m.File.Name()
		m.File.Close()
		os.Remove(n)
	}
}

func basicMessage(mainm *mail.Message, f *os.File) (*Message, error) {
	msg0 := &Message{}
	msg0.Header = mainm.Header
	msg0.Header = normalizeHeaders(msg0.Header)
	msg0.File = f
	if f != nil {
		msg0.Path = f.Name()
	}
	msg0.Kind = MSG_MESSAGE
	msg0.tempbody = mainm.Body

	tf2n, tf2, _ := tempFile()
	io.Copy(tf2, msg0.tempbody)

	if f != nil {
		fn := f.Name()
		f.Close()
		os.Remove(fn)
	}
	tf2.Seek(0, 0)
	msg0.File = tf2
	msg0.Path = tf2n

	cte := msg0.Header.Get("Content-Transfer-Encoding")
	cte = strings.TrimSpace(cte)
	cte = strings.ToLower(cte)

	if cte == "base64" && !msg0.decoded {
		bio := bufio.NewReader(msg0.File)
		tf3n, tf3, _ := tempFile()
		for {
			line, err := bio.ReadString('\n')
			if err != nil {
				break
			}
			if strings.HasSuffix(line, CRLF) {
				tf3.WriteString(line[:len(line)-2])
			} else {
				tf3.WriteString(line[:len(line)-1])
			}
		}
		tf3.Sync()
		tf3.Seek(0, 0)

		trdr := base64.NewDecoder(base64.StdEncoding, tf3)
		tf2.Seek(0, 0)
		tf2.Truncate(0)
		log.Println(io.Copy(tf2, trdr))
		tf2.Sync()
		tf2.Seek(0, 0)
		tf3.Close()
		os.Remove(tf3n)
		msg0.decoded = true
	} else if cte == "quoted-printable" && !msg0.decoded {
		bio := newQuotedPrintableReaderByCharsetId(-1)
		pn0 := header.MapParams(msg0.Header.Get("Content-Type"))
		if len(pn0["charset"]) > 0 {
			bio = newQuotedPrintableReaderByCharsetStr(pn0["charset"])
		}
		tf3n, tf3, _ := tempFile()
		bio.Decode(tf2, tf3)
		//io.Copy(tf3, bio)
		tf3.Sync()
		tf3.Seek(0, 0)
		tf2.Seek(0, 0)
		tf2.Truncate(0)
		io.Copy(tf2, tf3)
		tf2.Sync()
		tf2.Seek(0, 0)
		tf3.Close()
		os.Remove(tf3n)
		msg0.decoded = true
	}

	return msg0, nil
}

func alternativeMessage(mainm *mail.Message, f *os.File) (*Message, error) {
	a, b := multipartMessage(mainm, f)
	if b != nil {
		return &Message{Header: mainm.Header}, errors.New("alternativeMessage(mainm *mail.Message, f *os.File) (*Message, error): " + b.Error())
	}
	a.Kind = MSG_MULTIPARTALTERNATIVE
	return a, b
}

func multipartMessage(mainm *mail.Message, f *os.File) (*Message, error) {
	ct := mainm.Header.Get("Content-Type")
	log.Println("it's multipart mixed", ct)
	boundary, err := getBoundary(ct)
	log.Println("`" + boundary + "`")
	if err != nil {
		h2 := mainm.Header.Get("X-Invalid-Header")
		boundary, err = getBoundary(h2)
		if err != nil {
			log.Println("[multipartMessage] BOUNDARY ERROR", err)
			return nil, errors.New("BOUNDARY ERROR: " + err.Error())
		}
	} else {
		hn2 := mainm.Header.Get("X-Invalid-Header")
		log.Println("X-Invalid-Header", hn2)
		b2, err := getBoundary(hn2)
		if err == nil {
			log.Println("boundary was replaced from", boundary, "to", b2)
			boundary = b2
		}
	}
	boundary = strings.Trim(boundary, "\"")
	rdr := bufio.NewReader(mainm.Body)
	for {
		// read first boundary
		line, lerr := rdr.ReadString('\n')
		//log.Println(line)
		if lerr != nil {
			//log.Println("FGHG ERROR", lerr)
			if lerr.Error() == "EOF" {
				break
			} else {
				return nil, errors.New("line, lerr := rdr.ReadString('\n') NOT EOF: " + lerr.Error())
			}
		}
		if strings.HasPrefix(line, "--"+boundary) {
			break
		}
	}

	msg0 := &Message{}
	msg0.File = f
	if f != nil {
		msg0.Path = f.Name()
	}
	msg0.Kind = MSG_MULTIPARTMIXED
	msg0.Header = mainm.Header
	msg0.Children = make([]*Message, 0)

	bound1 := "--" + boundary
	bound2 := "--" + boundary + "--"

	cont0 := true

	for cont0 {
		_, tf, err := tempFile()
		if err != nil {
			return nil, errors.New("_, tf, err := tempFile(): " + err.Error())
		}
		for {
			line, err := rdr.ReadString('\n')
			if err != nil {
				break
			}
			if strings.HasPrefix(line, bound2) {
				cont0 = false
				break
			} else if strings.HasPrefix(line, bound1) {
				break
			}
			tf.WriteString(line)
		}
		tf.Seek(0, 0)
		//nmsg, err := readMessage(tf)
		nmsg, err := mail.ReadMessage(tf)
		if err != nil {
			log.Println("Error reading message!", err)
			log.Println("Native mail error!")
			//if nmsg == nil {
			return &Message{Header: mainm.Header}, err
			//}
			//mmmsg2 := &Message{}
			//mmmsg2.Header = nmsg.Header
			//log.Println(nmsg.Header)
			//if nmsg.Body == nil {
			//	err = errors.New("[mailmessage] EOF while trying to read the message body.")
			//	return mmmsg2, err
			//}
			//return mmmsg2, err
		}
		var mmmsg *Message
		contentType := nmsg.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/mixed") || strings.HasPrefix(contentType, "multipart/related") {
			mmmsg, err = multipartMessage(nmsg, tf)
		} else if strings.HasPrefix(contentType, "multipart/alternative") {
			mmmsg, err = alternativeMessage(nmsg, tf)
		} else {
			mmmsg, err = basicMessage(nmsg, tf)
		}
		if err != nil {
			log.Println("err johnson", err)
			return nil, err
		}
		msg0.Children = append(msg0.Children, mmmsg)
	}
	f.Close()
	os.Remove(msg0.Path)
	msg0.Path = ""
	msg0.File = nil
	return msg0, nil
}

func normalizeHeaders(h mail.Header) mail.Header {
	for k, v := range h {
		for k2 := range v {
			v[k2] = utf8b.ReplaceAllStringFunc(v[k2], func(str0 string) string {
				log.Println("UTF8-BINARY")
				vv := str0[10 : len(str0)-2]
				v8, err := base64.StdEncoding.DecodeString(vv)
				if err != nil {
					log.Println("FATAL B64 DECODE", err)
					return str0
				}
				return string(v8)
			})
			v[k2] = utf8q.ReplaceAllStringFunc(v[k2], func(str0 string) string {
				buff := new(bytes.Buffer)
				vv := str0[10 : len(str0)-2]
				buff.WriteString(vv)
				ior := newQuotedPrintableReaderByCharsetId(-1)
				buff2 := new(bytes.Buffer)
				ior.Decode(buff, buff2)
				return buff2.String()
			})
			v[k2] = iso88591q2.ReplaceAllStringFunc(v[k2], func(str0 string) string {
				buff := new(bytes.Buffer)
				vv := str0[15 : len(str0)-2]
				buff.WriteString(vv)
				ior := newQuotedPrintableReaderByCharsetStr("iso-8859-1")
				buff2 := new(bytes.Buffer)
				ior.Decode(buff, buff2)
				return buff2.String()
			})
			v[k2] = iso88591q.ReplaceAllStringFunc(v[k2], func(str0 string) string {
				buff := new(bytes.Buffer)
				vv := str0[14 : len(str0)-1]
				buff.WriteString(vv)
				ior := newQuotedPrintableReaderByCharsetStr("iso-8859-1")
				buff2 := new(bytes.Buffer)
				ior.Decode(buff, buff2)
				return buff2.String()
			})
			if strings.HasPrefix(v[k2], "\"") && strings.HasSuffix(v[k2], "\"") {
				v[k2] = strings.Trim(v[k2], "\"")
			}
		}
		h[k] = v
	}
	return h
}

func getBoundary(contentType string) (string, error) {
	kv := header.MapParams(contentType)
	if len(kv["boundary"]) > 0 {
		return kv["boundary"], nil
	}
	return "", errors.New("Boundary not found!")
}

/*func getBoundary(contentType string) (string, error) {
	strs := strings.Split(contentType, ";")
	for _, v := range strs {
		v = strings.TrimSpace(v)
		//TODO: fix this workaround
		if v[0] == 'B' {
			v = "b" + v[1:]
		}
		if strings.HasPrefix(v, "boundary=") {
			return v[9:], nil
		}
	}
	return "", errors.New("Boundary not found!")
}*/

var (
	tfi = 1
)

func tempFile() (string, *os.File, error) {
	tfi++
	bs := make([]byte, 6)
	rand.Read(bs)
	fn := hex.EncodeToString(bs) + "_" + strconv.FormatInt(time.Now().Unix(), 10) + "_" + strconv.Itoa(tfi) + ".dat"
	p0 := path.Join(os.TempDir(), fn)
	file, err := os.OpenFile(p0, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return p0, file, errors.New("file, err := os.OpenFile(p0, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666): " + err.Error())
	}
	return p0, file, nil
}
