package proto

import (
	"bytes"
	"reflect"
	"testing"

	. "chunkymonkey/types"
	te "testencoding"
)

func testPacketSerial(t *testing.T, inputPkt, outputPkt interface{}, expectedSerialization te.IBytesMatcher) {
	ps := new(PacketSerializer)

	// Test reading.
	input := new(bytes.Buffer)
	expectedSerialization.Write(input)
	if err := ps.ReadPacket(input, inputPkt); err != nil {
		t.Errorf("Unexpected error reading packet: %v", err)
	} else {
		if !reflect.DeepEqual(outputPkt, inputPkt) {
			t.Errorf("Packet did not read expected value:\n  expected: %#v\n    result: %#v", outputPkt, inputPkt)
		}
	}

	// Test writing.
	output := new(bytes.Buffer)
	if err := ps.WritePacket(output, outputPkt); err != nil {
		t.Errorf("Unexpected error writing packet: %v\n  %#v\v", err, outputPkt)
	} else {
		if err := te.Matches(expectedSerialization, output.Bytes()); err != nil {
			t.Errorf("Output of writing packet did not match: %v\n  %#v", err, outputPkt)
		}
	}
}

func TestPacketLogin(t *testing.T) {
	testPacketSerial(
		t,
		&PacketLogin{},
		&PacketLogin{
			VersionOrEntityId: 5,
			Username:          "username",
			MapSeed:           123,
			GameMode:          1,
			Dimension:         DimensionNormal,
			Difficulty:        GameDifficultyNormal,
			WorldHeight:       128,
			MaxPlayers:        12,
		},
		te.LiteralString(""+
			"\x00\x00\x00\x05"+ // Version/EntityID
			"\x00\x08\x00u\x00s\x00e\x00r\x00n\x00a\x00m\x00e"+ // Username
			"\x00\x00\x00\x00\x00\x00\x00\x7b"+ // MapSeed
			"\x00\x00\x00\x01"+ // GameMode
			"\x00"+ // Dimension
			"\x02"+ // Difficulty
			"\x80"+ // WorldHeight
			"\x0c"+ // MaxPlayers
			""),
	)
}
