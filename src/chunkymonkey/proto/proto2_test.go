package proto

import (
	"bytes"
	"reflect"
	"testing"

	. "chunkymonkey/types"
	te "testencoding"
)

const (
	Float32One = "\x3f\x80\x00\x00"
	Float32Two = "\x40\x00\x00\x00"

	Float64One   = "\x3f\xf0\x00\x00\x00\x00\x00\x00"
	Float64Two   = "\x40\x00\x00\x00\x00\x00\x00\x00"
	Float64Three = "\x40\x08\x00\x00\x00\x00\x00\x00"
	Float64Four  = "\x40\x10\x00\x00\x00\x00\x00\x00"
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

func Test_PacketLogin(t *testing.T) {
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
			"\x0c"), // MaxPlayers
	)
}

func Test_PacketUseEntity(t *testing.T) {
	testPacketSerial(
		t,
		&PacketUseEntity{},
		&PacketUseEntity{
			User:      2,
			Target:    5,
			LeftClick: true,
		},
		te.LiteralString(""+
			"\x00\x00\x00\x02"+
			"\x00\x00\x00\x05"+
			"\x01"),
	)
}

func Test_PacketPlayerPosition(t *testing.T) {
	testPacketSerial(
		t,
		&PacketPlayerPosition{},
		&PacketPlayerPosition{
			X: 1, Y: 2, Stance: 3, Z: 4,
			OnGround: true,
		},
		te.LiteralString(""+
			Float64One+
			Float64Two+
			Float64Three+
			Float64Four+
			"\x01"),
	)
}

func Test_PacketPlayerPositionLook(t *testing.T) {
	testPacketSerial(
		t,
		&PacketPlayerPositionLook{},
		&PacketPlayerPositionLook{
			X: 1, Y1: 2, Y2: 3, Z: 4,
			Look:     LookDegrees{Yaw: 1, Pitch: 2},
			OnGround: true,
		},
		te.LiteralString(""+
			Float64One+
			Float64Two+
			Float64Three+
			Float64Four+
			Float32One+
			Float32Two+
			"\x01"),
	)
}

func Test_PacketPlayerBlockInteract(t *testing.T) {
	testPacketSerial(
		t,
		&PacketPlayerBlockInteract{},
		&PacketPlayerBlockInteract{
			Position: BlockXyz{1, 2, 3},
			Face:     2,
			Tool: ItemSlot{
				ItemTypeId: 1,
				Count:      2,
				Data:       3,
			},
		},
		te.LiteralString(""+
			"\x00\x00\x00\x01"+
			"\x02"+
			"\x00\x00\x00\x03"+
			"\x02"+
			"\x00\x01"+
			"\x02"+
			"\x00\x03"),
	)

	// Test with last two fields missing (no tool used).
	testPacketSerial(
		t,
		&PacketPlayerBlockInteract{},
		&PacketPlayerBlockInteract{
			Position: BlockXyz{1, 2, 3},
			Face:     2,
			Tool: ItemSlot{
				ItemTypeId: -1,
			},
		},
		te.LiteralString(""+
			"\x00\x00\x00\x01"+
			"\x02"+
			"\x00\x00\x00\x03"+
			"\x02"+
			"\xff\xff"),
	)
}

func Test_PacketEntityMetadata(t *testing.T) {
	testPacketSerial(
		t,
		&PacketEntityMetadata{},
		&PacketEntityMetadata{
			EntityId: 5,
			Metadata: EntityMetadataTable{
				Items: []EntityMetadata{
					EntityMetadata{0, 0, byte(5)},
				},
			},
		},
		te.LiteralString(""+
			"\x00\x00\x00\x05"+
			"\x00\x05"+
			"\x7f"),
	)
}

func benchmarkPacket(b *testing.B, pkt interface{}) {
	output := bytes.NewBuffer(make([]byte, 0, 1024))
	var ps PacketSerializer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.WritePacket(output, pkt)
		output.Reset()
	}
}

func Benchmark_New_WritePacketLogin(b *testing.B) {
	benchmarkPacket(b, &PacketLogin{
		VersionOrEntityId: 5,
		Username:          "username",
		MapSeed:           123,
		GameMode:          1,
		Dimension:         DimensionNormal,
		Difficulty:        GameDifficultyNormal,
		WorldHeight:       128,
		MaxPlayers:        12,
	})
}

func Benchmark_Old_WritePacketLogin(b *testing.B) {
	output := bytes.NewBuffer(make([]byte, 0, 1024))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		commonWriteLogin(output, 5, "username", 123, 1, DimensionNormal, GameDifficultyNormal, 128, 12)
		output.Reset()
	}
}

func Benchmark_New_WritePacketKeepAlive(b *testing.B) {
	benchmarkPacket(b, &PacketKeepAlive{
		Id: 10,
	})
}

func Benchmark_Old_WritePacketKeepAlive(b *testing.B) {
	output := bytes.NewBuffer(make([]byte, 0, 1024))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		WriteKeepAlive(output, 10)
		output.Reset()
	}
}

func Benchmark_New_WritePacketEntityMetadata(b *testing.B) {
	benchmarkPacket(b, &PacketEntityMetadata{
		EntityId: 5,
		Metadata: EntityMetadataTable{
			Items: []EntityMetadata{
				EntityMetadata{0, 0, byte(5)},
			},
		},
	})
}
