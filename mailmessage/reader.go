package mailmessage

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/mail"
	//"net/textproto"
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

type Message struct {
	Kind     int
	Header   mail.Header
	tempbody io.Reader
	Path     string
	File     *os.File
	Children []*Message
	decoded  bool
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

func (m *Message) HTML() string {
	var rdr *os.File
	if m.Kind == MSG_MULTIPARTALTERNATIVE {
		// find html
		for k := range m.Children {
			ct := m.Children[k].Header.Get("Content-Type")
			if strings.HasPrefix(ct, "text/html") {
				rdr = m.Children[k].File
				log.Println("^^^ HTML HEADER:::", m.Children[k].Header)
				break
			}
		}
	} else if m.Kind == MSG_MESSAGE {
		ct := m.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "text/html") {
			rdr = m.File
		}
	}
	if rdr == nil {
		return ""
	}
	var buffer bytes.Buffer
	io.Copy(&buffer, rdr)
	rdr.Seek(0, 0)
	return buffer.String()
}

func (m *Message) Plaintext() string {
	var rdr io.Reader
	if m.Kind == MSG_MULTIPARTALTERNATIVE {
		// find text
		for k := range m.Children {
			ct := m.Children[k].Header.Get("Content-Type")
			if strings.HasPrefix(ct, "text/plain") {
				rdr = m.Children[k].File
				break
			}
		}
	} else if m.Kind == MSG_MESSAGE {
		ct := m.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "text/plain") {
			rdr = m.File
		}
	}
	if rdr == nil {
		return ""
	}
	var buffer bytes.Buffer
	io.Copy(&buffer, rdr)
	return buffer.String()
}

func New(rdr *bufio.Reader) (*Message, error) {

	line, err := rdr.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(line, "+OK") {
		log.Println("NEW MAIL", line[4:])
	} else if strings.HasPrefix(line, "-ERR") {
		return nil, errors.New(line[5:])
	} else {
		return nil, errors.New("Unknown pop3 server response `" + line + "`")
	}
	// save to a temporary file
	nowf := "msg_" + strconv.FormatInt(time.Now().Unix(), 10) + ".dat"
	fil0, err := os.OpenFile(path.Join(os.TempDir(), nowf), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)

	if err != nil {
		return nil, err
	}

	for {
		line, err = rdr.ReadString('\n')
		if err != nil {
			return nil, err
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
	fil0.Seek(0, 0)
	mainm, err := mail.ReadMessage(fil0)
	if err != nil {
		return nil, err
	}
	normalizeHeaders(mainm.Header)
	contentType := mainm.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/mixed") || strings.HasPrefix(contentType, "multipart/related") {
		return multipartMessage(mainm, fil0)
	} else if strings.HasPrefix(contentType, "multipart/alternative") {
		return alternativeMessage(mainm, fil0)
	}
	return basicMessage(mainm, fil0)
}

func basicMessage(mainm *mail.Message, f *os.File) (*Message, error) {
	msg0 := &Message{}
	msg0.Header = mainm.Header
	normalizeHeaders(msg0.Header)
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
		bio := newQuotedPrintableReader(msg0.File)
		tf3n, tf3, _ := tempFile()
		io.Copy(tf3, bio)
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
		return nil, b
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
		log.Println("BOUNDARY ERROR", err)
		return nil, err
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
				return nil, lerr
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
			return nil, err
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
		nmsg, err := mail.ReadMessage(tf)
		if err != nil {
			log.Println("nmsg, err := mail.ReadMessage(tf)", err)
			continue
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
			continue
		}
		msg0.Children = append(msg0.Children, mmmsg)
	}
	f.Close()
	os.Remove(msg0.Path)
	msg0.Path = ""
	msg0.File = nil
	return msg0, nil
}

func normalizeHeaders(h mail.Header) {
	utf8b := regexp.MustCompile(`=\?UTF-8\?B\?(.*?)\?=`)
	utf8q := regexp.MustCompile(`=\?UTF-8\?B\?(.*?)\?=`)
	iso88591q := regexp.MustCompile(`\?iso-8859-1\?Q\?(.*?)\?`)
	iso88591q2 := regexp.MustCompile(`=\?iso-8859-1\?Q\?(.*?)=\?`)
	for _, v := range h {
		for k2 := range v {
			v[k2] = utf8b.ReplaceAllStringFunc(v[k2], func(str0 string) string {
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
				ior := newQuotedPrintableReader(buff)
				buff2 := new(bytes.Buffer)
				io.Copy(buff2, ior)
				return buff2.String()
			})
			v[k2] = iso88591q2.ReplaceAllStringFunc(v[k2], func(str0 string) string {
				buff := new(bytes.Buffer)
				vv := str0[15 : len(str0)-2]
				buff.WriteString(vv)
				ior := newQuotedPrintableReader(buff)
				buff2 := new(bytes.Buffer)
				io.Copy(buff2, ior)
				return buff2.String()
			})
			v[k2] = iso88591q.ReplaceAllStringFunc(v[k2], func(str0 string) string {
				buff := new(bytes.Buffer)
				vv := str0[14 : len(str0)-1]
				buff.WriteString(vv)
				ior := newQuotedPrintableReader(buff)
				buff2 := new(bytes.Buffer)
				io.Copy(buff2, ior)
				return buff2.String()
			})
			v[k2] = strings.Trim(v[k2], "\"")
			/*if strings.HasPrefix(v[k2], "=?UTF-8?B?") {
				// UTF-8 BASE64
				vv := v[k2][10 : len(v[k2])-2]
				vv := v[k2][10 : len(v[k2])-2]
				v8, err := base64.StdEncoding.DecodeString(vv)
				if err != nil {
					log.Println("FATAL B64 DECODE", err)
				} else {
					v[k2] = string(v8)
				}
			} else if strings.HasPrefix(v[k2], "=?UTF-8?Q?") {
				buff := new(bytes.Buffer)
				vv := v[k2][10 : len(v[k2])-2]
				buff.WriteString(vv)
				ior := newQuotedPrintableReader(buff)
				buff2 := new(bytes.Buffer)
				io.Copy(buff2, ior)
				v[k2] = buff2.String()
			} else if strings.HasPrefix(v[k2], "?iso-8859-1?Q?") {
				buff := new(bytes.Buffer)
				vv := v[k2][14 : len(v[k2])-1]
				buff.WriteString(vv)
				ior := newQuotedPrintableReader(buff)
				buff2 := new(bytes.Buffer)
				io.Copy(buff2, ior)
				v[k2] = buff2.String()
			} else if strings.HasPrefix(v[k2], "?iso-8859-1?B?") {
				vv := v[k2][14 : len(v[k2])-1]
				v8, err := base64.StdEncoding.DecodeString(vv)
				if err != nil {
					log.Println("FATAL B64 DECODE", err)
				} else {
					v[k2] = string(v8)
				}
			}*/
		}
	}
}

func getBoundary(contentType string) (string, error) {
	strs := strings.Split(contentType, ";")
	for _, v := range strs {
		v = strings.TrimSpace(v)
		if strings.HasPrefix(v, "boundary=") {
			return v[9:], nil
		}
	}
	return "", errors.New("Boundary not found!")
}

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
	return p0, file, err
}
