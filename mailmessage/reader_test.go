package mailmessage

import (
	"bytes"
	"github.com/gabstv/latinx"
	"net/mail"
	"testing"
)

func TestQP(t *testing.T) {
	hdr := mail.Header{
		"From":  []string{"=?UTF-8?Q?ANT=C3=94NIA_FERNANDES_PESSOA_SENA?= <senna@gmail.com>"},
		"From2": []string{"=?UTF-8?Q?image=2Ejpg?=_0.jpg =?UTF-8?Q?image=2Ejpg?=_0.jpg"},
	}
	normalizeHeaders(hdr)
	if hdr.Get("From") != "ANTÔNIA_FERNANDES_PESSOA_SENA <senna@gmail.com>" {
		t.Errorf("%v != %v ", hdr.Get("From"), "ANTÔNIA_FERNANDES_PESSOA_SENA <senna@gmail.com>")
	}
	if hdr.Get("From2") != "image.jpg_0.jpg image.jpg_0.jpg" {
		t.Errorf("%v != %v ", hdr.Get("From2"), "image.jpg_0.jpg image.jpg_0.jpg")
	}
}
func TestQPIso(t *testing.T) {
	buff := new(bytes.Buffer)
	buff2 := new(bytes.Buffer)
	buff.WriteString(`Homem, =F3culos, =E7aba=E7o=EA!
1 =3D 2 - 1`)
	isordr := newQuotedPrintableReaderByCharsetId(latinx.Latin1)
	isordr.Decode(buff, buff2)
	if buff2.String() != "Homem, óculos, çabaçoê!\n1 = 2 - 1" {
		t.Errorf("`%v`\n!=\n`%v`\n", buff2.String(), "Homem, óculos, çabaçoê!\n1 = 2 - 1")
		t.Logf("(%X) (%X)", buff2.String(), "Homem, óculos, çabaçoê!\n1 = 2 - 1")
	}
}
