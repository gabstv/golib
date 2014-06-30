package mailmessage

import (
	"bytes"
	"github.com/gabstv/latinx"
	"testing"
)

func QP_TestIso(t *testing.T) {
	buff := new(bytes.Buffer)
	buff.WriteString(`Homem, =F3culos, =E7aba=E7o=EA!
1 =3D 2 - 1`)
	isordr := newQuotedPrintableReaderByCharsetId(latinx.Latin1)
	isordr.Decode(buff, buff)
	if buff.String() != "Homem, óculos, çabaçoê!\n1 = 2 - 1" {
		t.Errorf("%v != %v", buff.String(), "Homem, óculos, çabaçoê!\n1 = 2 - 1")
	}
}
