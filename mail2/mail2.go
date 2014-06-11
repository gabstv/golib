package mail2

import (
	"bufio"
	"bytes"
	"errors"
	//"fmt"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/mail"
	"strings"
)

var (
	boundary2 []byte
	boundary3 []byte
)

func ReadMessage(r io.Reader) (msg *Message, err error) {
	msg = &Message{}
	msg2, err := mail.ReadMessage(r)
	if err != nil {
		fmt.Println("LOCAO mail.ReadMessage(r)")
		return nil, err
	}
	msg.Header = msg2.Header
	reader := bufio.NewReader(msg2.Body)
	theBoundary, _, err := reader.ReadLine()
	boundary2 = make([]byte, len(theBoundary))
	if err != nil {
		fmt.Println("LOCAO reader.ReadLine()")
		return nil, err
	}
	for k, v := range theBoundary {
		boundary2[k] = v
	}
	isValid := true
	var buffer bytes.Buffer
	parts := 1
	container := make([]*mail.Message, 128)
	for isValid {
		line, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println("LOCAO LOOP reader.ReadLine()")
			return nil, err
		}
		//if len(line) < len(boundary)+3 {
		//fmt.Printf("LINE: %s\n", string(line))
		//}
		isb := isBoundary(boundary2, line)
		if isb < 1 {
			// [it's not the boundary] write to the current buffer
			buffer.Write(line)
			buffer.WriteString("\n")
		} else {
			// wrap it up
			cmsg, err := mail.ReadMessage(&buffer)
			if err != nil {
				fmt.Println("LOCAO WRAP! mail.ReadMessage(&buffer)")
				return nil, err
			}
			container[parts-1] = cmsg
			if isb == 2 {
				//fmt.Println("SHOULD END!")
				isValid = false
			} else {
				//fmt.Println("SHOULD START ANOTHER ONE!")
				parts++
				buffer.Truncate(0)
			}
		}
	}
	//
	curPart := 0
	foundAlternative := false
	for i := 0; i < parts; i++ {
		cmsg := container[i]
		ct := cmsg.Header["Content-Type"][0]
		ct = strings.Split(ct, ";")[0]
		ct = strings.TrimSpace(ct)
		//fmt.Println("Content-Type", ct)
		if ct == "multipart/alternative" {
			foundAlternative = true
			break
		}
	}
	if !foundAlternative {
		msg.Body = make([]*mail.Message, parts)
		msg.Attachments = make([]*mail.Message, 0)
		for i := 0; i < parts; i++ {
			msg.Body[i] = container[i]
		}
	} else {
		//fmt.Println("FOUND ALTERNATIVE!")
		msg.Attachments = make([]*mail.Message, parts-1)
		msg.Body = make([]*mail.Message, 1)
		// find the text one and add attachments
		for i := 0; i < parts; i++ {
			cmsg := container[i]
			ct := cmsg.Header["Content-Type"][0]
			ct = strings.Split(ct, ";")[0]
			ct = strings.TrimSpace(ct)
			//fmt.Println(ct)
			if ct == "multipart/alternative" {
				msg.Body[0] = cmsg
			} else {
				msg.Attachments[curPart] = cmsg
				curPart++
			}
		}
		// break alternative into text/plain and text/html
		if msg.Body[0] != nil {
			ct := msg.Body[0].Header["Content-Type"][0]
			cts := strings.Split(ct, ";")
			intnBondary := ""
			var intnBondaryBs []byte
			for _, v := range cts {
				v2 := strings.TrimSpace(v)
				if strings.HasPrefix(v2, "boundary") {
					// strip boundary= off
					v2 = strings.TrimPrefix(v2, "boundary")
					v2 = strings.TrimSpace(v2)
					v2 = strings.TrimPrefix(v2, "=")
					v2 = strings.TrimSpace(v2)
					intnBondary = "--" + v2
					//fmt.Println("TEXT BOUNDARY: ", intnBondary)
					intnBondaryBs = []byte(intnBondary)
					boundary3 = make([]byte, len(intnBondaryBs))
					for k, v := range intnBondaryBs {
						boundary3[k] = v
					}
					break
				}
			}
			isValid = true
			buffer.Truncate(0)
			parts = 1
			container = make([]*mail.Message, 64)
			reader = bufio.NewReader(msg.Body[0].Body)
			for isValid {
				line, _, er3 := reader.ReadLine()
				//fmt.Println("TEXT: ", string(line))
				if er3 != nil {
					return nil, er3
				}
				if len(line) > 2 && len(intnBondary) < 1 {
					intnBondary = string(line)
					//fmt.Println("TEXT% BOUNDARY: ", intnBondary)
					intnBondaryBs = []byte(intnBondary)
					boundary3 = make([]byte, len(intnBondaryBs))
					for k, v := range intnBondaryBs {
						boundary3[k] = v
					}
					continue
				}
				isb := isBoundary(boundary3, line)
				if isb < 1 {
					// [it's not the boundary] write to the current buffer
					buffer.Write(line)
					buffer.WriteString("\n")
				} else {
					if buffer.Len() < 2 && isb == 1 {
						continue
					}
					// wrap it up
					cmsg, err := mail.ReadMessage(&buffer)
					if err != nil {
						return nil, err
					}
					container[parts-1] = cmsg
					if isb == 2 {
						//fmt.Println("[TEXT] SHOULD END!")
						isValid = false
					} else {
						//fmt.Println("[TEXT] SHOULD START ANOTHER ONE!")
						parts++
						buffer.Truncate(0)
					}
				}
			}
			msg.Body = make([]*mail.Message, parts)
			for i := 0; i < parts; i++ {
				msg.Body[i] = container[i]
			}
		}
	}
	return msg, nil
}

func ReadMessageStr(r string) (msg *Message, err error) {
	var buffer bytes.Buffer
	buffer.WriteString(r)
	return ReadMessage(&buffer)
}

type Message struct {
	Header      mail.Header
	Body        []*mail.Message
	Attachments []*mail.Message
}

func (m *Message) GetTextMessage(preferHTML bool) (*mail.Message, error) {
	var msg *mail.Message

	for i := 0; i < len(m.Body); i++ {
		cmsg := m.Body[i]

		if msg == nil {
			msg = cmsg
			continue
		}

		ct := cmsg.Header["Content-Type"][0]
		ct = strings.Split(ct, ";")[0]
		ct = strings.TrimSpace(ct)

		if preferHTML && ct == "text/html" {
			msg = cmsg
			break
		} else if !preferHTML && ct == "text/plain" {
			msg = cmsg
			break
		}
	}

	if msg == nil {
		return nil, errors.New("This e-mail contains no body.")
	}

	return msg, nil
}

func (m *Message) GetTextBody(preferHTML bool) (string, error) {
	// try to get html first
	var buffer bytes.Buffer
	var msg *mail.Message
	var err error

	msg, err = m.GetTextMessage(preferHTML)

	if msg == nil {
		return "", err
	}

	io.Copy(&buffer, msg.Body)

	return buffer.String(), nil
}

func (m *Message) GetAttachmentBytes(index int) (file []byte, md5checksum [16]byte, err error) {
	file = nil
	err = nil
	if len(m.Attachments) <= index {
		err = errors.New("Index overflow.")
		return
	}
	cmsg := m.Attachments[index]
	rdr := bufio.NewReader(cmsg.Body)
	var buffer bytes.Buffer
	for {
		line, err2 := rdr.ReadString('\n')
		if len(line) > 0 {
			lineb, err3 := base64.StdEncoding.DecodeString(line)
			if err3 != nil {
				err = err3
				return
			}
			buffer.Write(lineb)
		}
		if err2 != nil {
			break
		}
	}
	file = buffer.Bytes()
	md5checksum = md5.Sum(file)
	return
}

func isBoundary(boundary, line []byte) int {
	if len(line) < len(boundary) {
		return 0
	}
	if len(line) == len(boundary)+2 {
		//fmt.Println("BOUNDARY ", boundary)
		//fmt.Println("ORIGINAL BOUNDARY ", boundary2)
		//fmt.Println("MAGIC LINE ", line, line[len(line)-1], line[len(line)-2])
		if compareBytes(boundary, line) {
			if line[len(line)-2] == byte(45) && line[len(line)-1] == byte(45) {
				return 2
			}
		}
	} else if len(line) == len(boundary) {
		if compareBytes(boundary, line) {
			return 1
		}
	}
	return 0
}

func compareBytes(lineA, lineB []byte) bool {
	len0 := len(lineA)
	if len0 > len(lineB) {
		len0 = len(lineB)
	}
	for i := 0; i < len0; i++ {
		if lineA[i] != lineB[i] {
			return false
		}
	}
	return true
}
