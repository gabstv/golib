package mailmessage

import (
	"net/mail"
	"testing"
)

func testQP(t *testing.T) {
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
