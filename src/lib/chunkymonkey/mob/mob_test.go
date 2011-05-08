package mob

import (
	"bytes"
	"testing"

	"chunkymonkey/types"
)

func TestCreeperSpawn(t *testing.T) {
		want := "\x18\x00\x00\x00\x002\x00\x00\x01o\x00\x00\b\xc0\xff\xff\ua4ce\x00\x00\x01\x11\x01\x7f"
		m := NewCreeper()
		m.CreeperSetBlueAura()
		m.SetBurning()
		m.SetPosition(types.AbsXyz{11, 70, -172})
		buf := bytes.NewBuffer(nil)
		if err := m.SendSpawn(buf); err != nil {
			t.Fatal(err)
		}
		if buf.String() != want {
			t.Errorf("Resulting raw data mismatch, wanted:\n\t%q\n\tbut got: \n\t%q", want, buf.String())
		}
}
