package mailmessage

import (
	"net/mail"
	"testing"
)

func testQP(t *testing.T) {
	hdr := mail.Header{"From": []string{"=?UTF-8?Q?ANT=C3=94NIA_FERNANDES_PESSOA_SENA?= <senna@gmail.com>"}}
	normalizeHeaders(hdr)
	if hdr.Get("From") != "ANTÔNIA_FERNANDES_PESSOA_SENA <senna@gmail.com>" {
		t.Errorf("%v != %v ", hdr.Get("From"), "ANTÔNIA_FERNANDES_PESSOA_SENA <senna@gmail.com>")
	}
}
