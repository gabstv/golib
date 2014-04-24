package mail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/smtp"
)

type Address struct {
	Email string
	Name  string
}

type Message struct {
	From          Address
	To            []Address
	Subject       string
	HTMLBody      string
	PlaintextBody string
	Files         *AttachmentList
}

func NewMessage() *Message {
	m := &Message{}
	m.Files = NewAttachmentList()
	return m
}

func (m *Message) AddRecipient(to Address) {
	if m.To == nil {
		m.To = make([]Address, 0)
	}
	m.To = append(m.To, to)
}

type SMTP struct {
	Auth    smtp.Auth
	Address string
}

func NewSMTP(auth smtp.Auth, address string) *SMTP {
	return &SMTP{auth, address}
}

func (s *SMTP) SubmitHTML(fromEmail, fromName, toEmail, toName, subject, htmlBody string, files *AttachmentList) error {
	m := NewMessage()
	m.From.Name = fromName
	m.From.Email = fromEmail
	m.AddRecipient(Address{toEmail, toName})
	m.HTMLBody = htmlBody
	m.Files = files
	return s.submit(m)
}

func (s *SMTP) SubmitPlaintext(fromEmail, fromName, toEmail, toName, subject, plainBody string, files *AttachmentList) error {
	m := NewMessage()
	m.From.Name = fromName
	m.From.Email = fromEmail
	m.AddRecipient(Address{toEmail, toName})
	m.PlaintextBody = plainBody
	m.Files = files
	return s.submit(m)
}

func (s *SMTP) SubmitMixed(fromEmail, fromName, toEmail, toName, subject, plainBody string, htmlBody string, files *AttachmentList) error {
	m := NewMessage()
	m.From.Name = fromName
	m.From.Email = fromEmail
	m.AddRecipient(Address{toEmail, toName})
	m.HTMLBody = htmlBody
	m.PlaintextBody = plainBody
	m.Files = files
	return s.submit(m)
}

func (s *SMTP) writeTextEmailPart(w io.Writer, contentType, body string) (n int, err error) {
	var n2 int
	// write content-type
	n2, err = w.Write([]byte("Content-Type: " + contentType + "\r\n"))
	n += n2
	if err != nil {
		return
	}
	// write content transfer encoding
	n2, err = w.Write([]byte("Content-Transfer-Encoding: 8bit\r\n"))
	n += n2
	if err != nil {
		return
	}
	// write body
	n2, err = w.Write([]byte(fmt.Sprintf("\r\n%s\r\n", body)))
	n += n2
	return
}

func (s *SMTP) submit(msg *Message) error {
	var buffer bytes.Buffer
	bmarker := newBoundary()

	//// [START] [HEADERS] write the email headers
	// FROM
	buffer.WriteString(fmt.Sprintf("From: %s <%s>\r\n", msg.From.Name, msg.From.Email))
	// TO
	var tol bytes.Buffer
	for i := 0; i < len(msg.To); i++ {
		if i > 0 {
			tol.WriteString(", ")
		}
		tol.WriteString(msg.To[i].Name)
		tol.WriteString(" <")
		tol.WriteString(msg.To[i].Email)
		tol.WriteString(">")
	}
	buffer.WriteString(fmt.Sprintf("To: %s\r\n", tol.String()))
	// SUBJECT
	buffer.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	// MIME-Version
	buffer.WriteString("MIME-Version: 1.0\r\n")
	// Content-Type
	buffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n--%s\r\n", bmarker, bmarker))
	// [END] [HEADERS]

	// Write HTML (if any)
	if len(msg.HTMLBody) > 0 {
		s.writeTextEmailPart(buffer, "text/html; charset=UTF-8", msg.HTMLBody)
	}
	// Write Plaintext (if any)
	if len(msg.PlaintextBody) > 0 {
		if len(msg.HTMLBody) > 0 {
			// boundary / marker / separator
			buffer.WriteString(fmt.Sprintf("--%s\r\n", bmarker))
		}
		s.writeTextEmailPart(buffer, "text/plain; charset=UTF-8", msg.PlaintextBody)
	}

	tols := make([]string, lwn(msg.To))
	for k, v := range msg.To {
		tols[k] = v.Email
	}

	if msg.Files == nil {
		buffer.WriteString(fmt.Sprintf("--%s--", bmarker))
		return smtp.SendMail(s.Address, s.Auth, msg.From.Email, tols, buffer.Bytes())
	}

	if msg.Files.Count() < 1 {
		buffer.WriteString(fmt.Sprintf("--%s--", bmarker))
		return smtp.SendMail(s.Address, s.Auth, msg.From.Email, tols, buffer.Bytes())
	}

	var fbuff bytes.Buffer
	for curItem := files.First(); curItem != nil; curItem = curItem.Next() {

		fbuff.Truncate(0)
		//read and encode attachment
		content, _ := ioutil.ReadAll(curItem.Value.File)
		encoded := base64.StdEncoding.EncodeToString(content)

		//split the encoded file in lines (doesn't matter, but low enough not to hit a max limit)
		lineMaxLength := 500
		nbrLines := len(encoded) / lineMaxLength

		//append lines to buffer
		for i := 0; i < nbrLines; i++ {
			fbuff.WriteString(encoded[i*lineMaxLength:(i+1)*lineMaxLength] + "\n")
		}

		//part 3 will be the attachment
		buffer.WriteString(fmt.Sprintf("\r\nContent-Type: %s; name=\"%s\"\r\nContent-Transfer-Encoding:base64\r\nContent-Disposition: attachment; filename=\"%s\"\r\n\r\n%s\r\n--%s", curItem.Value.MimeType, curItem.Value.Name, curItem.Value.Name, fbuff.String(), bmarker))
		curItem.Value.File.Close()
	}
	fbuff.Truncate(0)
	buffer.WriteString("--")
	return smtp.SendMail(s.Address, s.Auth, msg.From.Email, tols, buffer.Bytes())
}
