package mail

import (
	"bytes"
	"fmt"
	"github.com/gabstv/golib/smtp2"
	"github.com/sloonz/go-mime-message"
	"github.com/sloonz/go-qprintable"
	"net/smtp"
	"os"
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

func (s *SMTP) SubmitHTML(fromEmail, fromName, toEmail, toName, subject, htmlBody string, files *AttachmentList) (int, error) {
	m := NewMessage()
	m.From.Name = fromName
	m.From.Email = fromEmail
	m.AddRecipient(Address{toEmail, toName})
	m.HTMLBody = htmlBody
	m.Files = files
	m.Subject = subject
	return s.submit(m)
}

func (s *SMTP) SubmitPlaintext(fromEmail, fromName, toEmail, toName, subject, plainBody string, files *AttachmentList) (int, error) {
	m := NewMessage()
	m.From.Name = fromName
	m.From.Email = fromEmail
	m.AddRecipient(Address{toEmail, toName})
	m.PlaintextBody = plainBody
	m.Files = files
	m.Subject = subject
	return s.submit(m)
}

func (s *SMTP) SubmitMixed(fromEmail, fromName, toEmail, toName, subject, plainBody string, htmlBody string, files *AttachmentList) (int, error) {
	m := NewMessage()
	m.From.Name = fromName
	m.From.Email = fromEmail
	m.AddRecipient(Address{toEmail, toName})
	m.HTMLBody = htmlBody
	m.PlaintextBody = plainBody
	m.Files = files
	m.Subject = subject
	return s.submit(m)
}

func (s *SMTP) SubmitMessage(msg *Message) error {
	_, err := s.submit(msg)
	return err
}

func (s *SMTP) submit(msg *Message) (int, error) {
	//var buffer bytes.Buffer
	bmarker := newBoundary()
	multipartmessage := message.NewMultipartMessage("mixed", bmarker)

	multipartmessage.SetHeader("From", fmt.Sprintf("%s <%s>", message.EncodeWord(msg.From.Name), msg.From.Email))
	multipartmessage.SetHeader("Return-Path", fmt.Sprintf("<%s>", msg.From.Email))
	// TO
	var tol bytes.Buffer
	for i := 0; i < len(msg.To); i++ {
		if i > 0 {
			tol.WriteString(", ")
		}
		tol.WriteString(message.EncodeWord(msg.To[i].Name))
		tol.WriteString(" <")
		tol.WriteString(msg.To[i].Email)
		tol.WriteString(">")
	}
	multipartmessage.SetHeader("To", tol.String())
	// SUBJECT
	multipartmessage.SetHeader("Subject", message.EncodeWord(msg.Subject))
	//// MIME-Version
	multipartmessage.SetHeader("MIME-Version", "1.0")
	//// [END] [HEADERS]

	if len(msg.HTMLBody) > 0 && len(msg.PlaintextBody) > 0 {
		alternatives := message.NewMultipartMessage("alternative", newBoundary())
		hmsg := message.NewTextMessage(qprintable.UnixTextEncoding, bytes.NewBufferString(msg.HTMLBody))
		hmsg.SetHeader("Content-Type", "text/html; charset=UTF-8")
		tmsg := message.NewTextMessage(qprintable.UnixTextEncoding, bytes.NewBufferString(msg.PlaintextBody))
		tmsg.SetHeader("Content-Type", "text/plain; charset=UTF-8")
		multipartmessage.AddPart(&alternatives.Message)
	} else if len(msg.HTMLBody) > 0 {
		hmsg := message.NewTextMessage(qprintable.UnixTextEncoding, bytes.NewBufferString(msg.HTMLBody))
		hmsg.SetHeader("Content-Type", "text/html; charset=UTF-8")
		multipartmessage.AddPart(hmsg)
	} else if len(msg.PlaintextBody) > 0 {
		tmsg := message.NewTextMessage(qprintable.UnixTextEncoding, bytes.NewBufferString(msg.PlaintextBody))
		tmsg.SetHeader("Content-Type", "text/plain; charset=UTF-8")
		multipartmessage.AddPart(tmsg)
	}

	tols := make([]string, len(msg.To))
	for k, v := range msg.To {
		tols[k] = v.Email
	}
	//
	if msg.Files == nil {
		goto Submit
	}
	if msg.Files.Count() < 1 {
		goto Submit
	}

	for curItem := msg.Files.First(); curItem != nil; curItem = curItem.Next() {
		if len(curItem.Value.Path) > 0 && curItem.Value.File == nil {
			curItem.Value.File, _ = os.Open(curItem.Value.Path)
			//TODO: treat this error
		}
		msg00 := message.NewBinaryMessage(curItem.Value.File)
		msg00.SetHeader("Content-Type", fmt.Sprintf("%v; name=\"%v\"", curItem.Value.MimeType, message.EncodeWord(curItem.Value.Name)))
		msg00.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", message.EncodeWord(curItem.Value.Name)))
		multipartmessage.AddPart(msg00)
	}

Submit:
	return 1024 * 1024, smtp2.SendMail(s.Address, s.Auth, msg.From.Email, tols, multipartmessage)
}
