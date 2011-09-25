package proto

import (
	"encoding/binary"
	"io"
	"log"
	"math"
	"os"
	"reflect"

	. "chunkymonkey/types"
)

type PacketLogin struct {
	VersionOrEntityId int32
	Username          string
	MapSeed           RandomSeed
	GameMode          int32
	Dimension         DimensionId
	Difficulty        GameDifficulty
	WorldHeight       byte
	MaxPlayers        byte
}

var (
	ErrorPacketNotPtr = os.NewError("packet not passed as a pointer")
	ErrorPacketNil = os.NewError("packet was passed by a nil pointer")
	ErrorStrLengthNegative = os.NewError("string length was negative")
	ErrorStrTooLong = os.NewError("string was too long")
	ErrorInternal = os.NewError("implementation problem with packetization")
)

type PacketSerializer struct {

	// Scratch space to be able to encode up to 64bit values without allocating.
	scratch [8]byte
}

func (ps *PacketSerializer) ReadPacket(reader io.Reader, packet interface{}) (err os.Error) {
	// TODO Check packet is CanSettable? (if settable at the top, does that
	// follow for all its descendants?)
	value := reflect.ValueOf(packet)
	kind := value.Kind()
	if kind != reflect.Ptr {
		return ErrorPacketNotPtr
	} else if value.IsNil() {
		return ErrorPacketNil
	}

	return ps.readData(reader, reflect.Indirect(value))
}

func (ps *PacketSerializer) readData(reader io.Reader, value reflect.Value) (err os.Error) {
	kind := value.Kind()

	switch kind {
	case reflect.Struct:
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			field := value.Field(i)
			ps.readData(reader, field)
		}

		// Integer types:

	case reflect.Int8:
		if _, err = io.ReadFull(reader, ps.scratch[0:1]); err != nil {
			return
		}
		value.SetInt(int64(ps.scratch[0]))
	case reflect.Int16:
		if _, err = io.ReadFull(reader, ps.scratch[0:2]); err != nil {
			return
		}
		value.SetInt(int64(binary.BigEndian.Uint16(ps.scratch[0:2])))
	case reflect.Int32:
		if _, err = io.ReadFull(reader, ps.scratch[0:4]); err != nil {
			return
		}
		value.SetInt(int64(binary.BigEndian.Uint32(ps.scratch[0:4])))
	case reflect.Int64:
		if _, err = io.ReadFull(reader, ps.scratch[0:8]); err != nil {
			return
		}
		value.SetInt(int64(binary.BigEndian.Uint64(ps.scratch[0:8])))
	case reflect.Uint8:
		if _, err = io.ReadFull(reader, ps.scratch[0:1]); err != nil {
			return
		}
		value.SetUint(uint64(ps.scratch[0]))
	case reflect.Uint16:
		if _, err = io.ReadFull(reader, ps.scratch[0:2]); err != nil {
			return
		}
		value.SetUint(uint64(binary.BigEndian.Uint16(ps.scratch[0:2])))
	case reflect.Uint32:
		if _, err = io.ReadFull(reader, ps.scratch[0:4]); err != nil {
			return
		}
		value.SetUint(uint64(binary.BigEndian.Uint32(ps.scratch[0:4])))
	case reflect.Uint64:
		if _, err = io.ReadFull(reader, ps.scratch[0:8]); err != nil {
			return
		}
		value.SetUint(binary.BigEndian.Uint64(ps.scratch[0:8]))

	case reflect.String:
		// TODO Maybe the tag field could/should suggest a max length.
		if _, err = io.ReadFull(reader, ps.scratch[0:2]); err != nil {
			return
		}
		length := int16(binary.BigEndian.Uint16(ps.scratch[0:2]))
		if length < 0 {
			return ErrorStrLengthNegative
		}
		codepoints := make([]uint16, length)
		if err = binary.Read(reader, binary.BigEndian, codepoints); err != nil {
			return
		}
		value.SetString(encodeUtf8(codepoints))

	default:
		// TODO
		typ := value.Type()
		log.Printf("Unimplemented type in packet: %v", typ)
		return ErrorInternal
	}
	return
}

func (ps *PacketSerializer) WritePacket(writer io.Writer, packet interface{}) (err os.Error) {
	value := reflect.ValueOf(packet)
	kind := value.Kind()
	if kind == reflect.Ptr {
		value = reflect.Indirect(value)
	}

	return ps.writeData(writer, value)
}

func (ps *PacketSerializer) writeData(writer io.Writer, value reflect.Value) (err os.Error) {
	kind := value.Kind()

	switch kind {
	case reflect.Struct:
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			field := value.Field(i)
			ps.writeData(writer, field)
		}

		// Integer types:

	case reflect.Int8:
		ps.scratch[0] = byte(value.Int())
		_, err = writer.Write(ps.scratch[0:1])
	case reflect.Int16:
		binary.BigEndian.PutUint16(ps.scratch[0:2], uint16(value.Int()))
		_, err = writer.Write(ps.scratch[0:2])
	case reflect.Int32:
		binary.BigEndian.PutUint32(ps.scratch[0:4], uint32(value.Int()))
		_, err = writer.Write(ps.scratch[0:4])
	case reflect.Int64:
		binary.BigEndian.PutUint64(ps.scratch[0:8], uint64(value.Int()))
		_, err = writer.Write(ps.scratch[0:8])
	case reflect.Uint8:
		ps.scratch[0] = byte(value.Uint())
		_, err = writer.Write(ps.scratch[0:1])
	case reflect.Uint16:
		binary.BigEndian.PutUint16(ps.scratch[0:2], uint16(value.Uint()))
		_, err = writer.Write(ps.scratch[0:2])
	case reflect.Uint32:
		binary.BigEndian.PutUint32(ps.scratch[0:4], uint32(value.Uint()))
		_, err = writer.Write(ps.scratch[0:4])
	case reflect.Uint64:
		binary.BigEndian.PutUint64(ps.scratch[0:8], value.Uint())
		_, err = writer.Write(ps.scratch[0:8])

	case reflect.String:
		lengthInt := value.Len()
		if lengthInt > math.MaxInt16 {
			return ErrorStrTooLong
		}
		binary.BigEndian.PutUint16(ps.scratch[0:2], uint16(lengthInt))
		if _, err = writer.Write(ps.scratch[0:2]); err != nil {
			return
		}
		codepoints := decodeUtf8(value.String())
		err = binary.Write(writer, binary.BigEndian, codepoints)

	default:
		// TODO
		typ := value.Type()
		log.Printf("Unimplemented type in packet: %v", typ)
		return ErrorInternal
	}

	return
}
