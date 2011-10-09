package proto

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"math"
	"reflect"
)

// Possible error values for reading and writing packets.
var (
	ErrorPacketNotPtr      = os.NewError("packet not passed as a pointer")
	ErrorUnknownPacketType = os.NewError("unknown packet type")
	ErrorUnexpectedPacket  = os.NewError("unexpected packet id")
	ErrorPacketNil         = os.NewError("packet was passed by a nil pointer")
	ErrorLengthNegative    = os.NewError("length was negative")
	ErrorStrTooLong        = os.NewError("string was too long")
	ErrorBadPacketData     = os.NewError("packet data well-formed but contains out of range values")
	ErrorBadChunkDataSize  = os.NewError("map chunk data length mismatches with size")
	ErrorMismatchingValues = os.NewError("packet data contains mismatching values")
	ErrorInternal          = os.NewError("implementation problem with packetization")
)

var (
	// Space to read unwanted data into. As the contents of this aren't used, it
	// doesn't require syncronization.
	dump [4096]byte
)

// IMinecraftMarshaler is the interface by which packet fields (or potentially
// even whole packets) can customize their serialization. It will only work for
// struct and slice-based types currently, as a hacky method of optimizing
// which packet fields are checked for this property.
type IMarshaler interface {
	MinecraftUnmarshal(reader io.Reader, ps *PacketSerializer) (err os.Error)
	MinecraftMarshal(writer io.Writer, ps *PacketSerializer) (err os.Error)
}

// PacketSerializer reads and writes packets. It is not safe to use one
// simultaneously between multiple goroutines.
//
// It does not take responsibility for reading/writing the packet ID byte
// header.
//
// It is designed to read and write struct types, and can only handle a few
// types - it is not a generalized serialization mechanism and isn't intended
// to be one. It exercises the freedom of having only limited types of packet
// structure partly for simplicity, and partly to allow for optimizations.
type PacketSerializer struct {
	// Scratch space to be able to encode up to 64bit values without allocating.
	scratch [8]byte
}

func (ps *PacketSerializer) ReadPacket(reader io.Reader, fromClient bool) (packet interface{}, err os.Error) {
	// Read packet ID.
	if _, err = io.ReadFull(reader, ps.scratch[0:1]); err != nil {
		return
	}

	pktInfo := &pktIdInfo[ps.scratch[0]]
	if !pktInfo.validPacket {
		return nil, ErrorUnknownPacketType
	}

	var expected bool
	if fromClient {
		expected = pktInfo.clientToServer
	} else {
		expected = pktInfo.serverToClient
	}
	if !expected {
		return nil, ErrorUnexpectedPacket
	}

	value := reflect.New(pktInfo.pktType)
	if err = ps.readData(reader, reflect.Indirect(value)); err != nil {
		return
	}

	return value.Interface(), nil
}

func (ps *PacketSerializer) readData(reader io.Reader, value reflect.Value) (err os.Error) {
	kind := value.Kind()

	switch kind {
	case reflect.Struct:
		valuePtr := value.Addr()
		if valueMarshaller, ok := valuePtr.Interface().(IMarshaler); ok {
			// Get the value to read itself.
			return valueMarshaller.MinecraftUnmarshal(reader, ps)
		}

		numField := value.NumField()
		for i := 0; i < numField; i++ {
			field := value.Field(i)
			if err = ps.readData(reader, field); err != nil {
				return
			}
		}

	case reflect.Slice:
		valuePtr := value.Addr()
		if valueMarshaller, ok := valuePtr.Interface().(IMarshaler); ok {
			// Get the value to read itself.
			return valueMarshaller.MinecraftUnmarshal(reader, ps)
		} else {
			return ErrorInternal
		}

	case reflect.Bool:
		if v, err := ps.readBool(reader); err != nil {
			return err
		} else {
			value.SetBool(v)
		}

		// Integer types:

	case reflect.Int8:
		if v, err := ps.readUint8(reader); err != nil {
			return err
		} else {
			value.SetInt(int64(int8(v)))
		}
	case reflect.Int16:
		if v, err := ps.readUint16(reader); err != nil {
			return err
		} else {
			value.SetInt(int64(int16(v)))
		}
	case reflect.Int32:
		if v, err := ps.readUint32(reader); err != nil {
			return err
		} else {
			value.SetInt(int64(int32(v)))
		}
	case reflect.Int64:
		if v, err := ps.readUint64(reader); err != nil {
			return err
		} else {
			value.SetInt(int64(v))
		}
	case reflect.Uint8:
		if v, err := ps.readUint8(reader); err != nil {
			return err
		} else {
			value.SetUint(uint64(v))
		}
	case reflect.Uint16:
		if v, err := ps.readUint16(reader); err != nil {
			return err
		} else {
			value.SetUint(uint64(v))
		}
	case reflect.Uint32:
		if v, err := ps.readUint32(reader); err != nil {
			return err
		} else {
			value.SetUint(uint64(v))
		}
	case reflect.Uint64:
		if v, err := ps.readUint64(reader); err != nil {
			return err
		} else {
			value.SetUint(v)
		}

		// Floating point types:

	case reflect.Float32:
		if v, err := ps.readFloat32(reader); err != nil {
			return err
		} else {
			value.SetFloat(float64(v))
		}
	case reflect.Float64:
		if v, err := ps.readFloat64(reader); err != nil {
			return err
		} else {
			value.SetFloat(v)
		}

	case reflect.String:
		// TODO Maybe the tag field could/should suggest a max length.
		lengthUint16, err := ps.readUint16(reader)
		if err != nil {
			return
		}
		length := int16(lengthUint16)
		if length < 0 {
			return ErrorLengthNegative
		}
		codepoints := make([]uint16, length)
		if err = binary.Read(reader, binary.BigEndian, codepoints); err != nil {
			return
		}
		value.SetString(encodeUtf8(codepoints))

	default:
		typ := value.Type()
		log.Printf("Unimplemented type in packet: %v", typ)
		return ErrorInternal
	}
	return
}

func (ps *PacketSerializer) WritePacket(writer io.Writer, packet interface{}) (err os.Error) {
	value := reflect.Indirect(reflect.ValueOf(packet))
	pktType := value.Type()

	// Write packet ID.
	var ok bool
	if ps.scratch[0], ok = pktTypeId[pktType]; !ok {
		return ErrorUnknownPacketType
	}
	if _, err = writer.Write(ps.scratch[0:1]); err != nil {
		return
	}

	return ps.writeData(writer, value)
}

func (ps *PacketSerializer) writeData(writer io.Writer, value reflect.Value) (err os.Error) {
	kind := value.Kind()

	switch kind {
	case reflect.Struct:
		valuePtr := value.Addr()
		if valueMarshaller, ok := valuePtr.Interface().(IMarshaler); ok {
			// Get the value to write itself.
			return valueMarshaller.MinecraftMarshal(writer, ps)
		}

		numField := value.NumField()
		for i := 0; i < numField; i++ {
			field := value.Field(i)
			if err = ps.writeData(writer, field); err != nil {
				return
			}
		}

	case reflect.Slice:
		valuePtr := value.Addr()
		if valueMarshaller, ok := valuePtr.Interface().(IMarshaler); ok {
			// Get the value to write itself.
			return valueMarshaller.MinecraftMarshal(writer, ps)
		} else {
			return ErrorInternal
		}

	case reflect.Bool:
		err = ps.writeBool(writer, value.Bool())

		// Integer types:

	case reflect.Int8:
		err = ps.writeUint8(writer, uint8(int8(value.Int())))
	case reflect.Int16:
		err = ps.writeUint16(writer, uint16(int16(value.Int())))
	case reflect.Int32:
		err = ps.writeUint32(writer, uint32(int32(value.Int())))
	case reflect.Int64:
		err = ps.writeUint64(writer, uint64(int64(value.Int())))
	case reflect.Uint8:
		err = ps.writeUint8(writer, uint8(value.Uint()))
	case reflect.Uint16:
		err = ps.writeUint16(writer, uint16(value.Uint()))
	case reflect.Uint32:
		err = ps.writeUint32(writer, uint32(value.Uint()))
	case reflect.Uint64:
		err = ps.writeUint64(writer, value.Uint())

		// Floating point types:

	case reflect.Float32:
		err = ps.writeFloat32(writer, float32(value.Float()))
	case reflect.Float64:
		err = ps.writeFloat64(writer, value.Float())

	case reflect.String:
		lengthInt := value.Len()
		if lengthInt > math.MaxInt16 {
			return ErrorStrTooLong
		}
		if err = ps.writeUint16(writer, uint16(lengthInt)); err != nil {
			return
		}
		codepoints := decodeUtf8(value.String())
		err = binary.Write(writer, binary.BigEndian, codepoints)

	default:
		typ := value.Type()
		log.Printf("Unimplemented type in packet: %v", typ)
		return ErrorInternal
	}

	return
}

// read/write bool.
func (ps *PacketSerializer) readBool(reader io.Reader) (v bool, err os.Error) {
	vUint8, err := ps.readUint8(reader)
	return vUint8 != 0, err
}
func (ps *PacketSerializer) writeBool(writer io.Writer, v bool) (err os.Error) {
	if v {
		return ps.writeUint8(writer, 1)
	}
	return ps.writeUint8(writer, 0)
}

// read/write uint8.
func (ps *PacketSerializer) readUint8(reader io.Reader) (v uint8, err os.Error) {
	if _, err = io.ReadFull(reader, ps.scratch[0:1]); err != nil {
		return
	}
	return ps.scratch[0], nil
}
func (ps *PacketSerializer) writeUint8(writer io.Writer, v uint8) (err os.Error) {
	ps.scratch[0] = v
	_, err = writer.Write(ps.scratch[0:1])
	return
}

// read/write uint16.
func (ps *PacketSerializer) readUint16(reader io.Reader) (v uint16, err os.Error) {
	if _, err = io.ReadFull(reader, ps.scratch[0:2]); err != nil {
		return
	}
	return binary.BigEndian.Uint16(ps.scratch[0:2]), nil
}
func (ps *PacketSerializer) writeUint16(writer io.Writer, v uint16) (err os.Error) {
	binary.BigEndian.PutUint16(ps.scratch[0:2], v)
	_, err = writer.Write(ps.scratch[0:2])
	return
}

// read/write uint32.
func (ps *PacketSerializer) readUint32(reader io.Reader) (v uint32, err os.Error) {
	if _, err = io.ReadFull(reader, ps.scratch[0:4]); err != nil {
		return
	}
	return binary.BigEndian.Uint32(ps.scratch[0:4]), nil
}
func (ps *PacketSerializer) writeUint32(writer io.Writer, v uint32) (err os.Error) {
	binary.BigEndian.PutUint32(ps.scratch[0:4], v)
	_, err = writer.Write(ps.scratch[0:4])
	return
}

// read/write uint64.
func (ps *PacketSerializer) readUint64(reader io.Reader) (v uint64, err os.Error) {
	if _, err = io.ReadFull(reader, ps.scratch[0:8]); err != nil {
		return
	}
	return binary.BigEndian.Uint64(ps.scratch[0:8]), nil
}
func (ps *PacketSerializer) writeUint64(writer io.Writer, v uint64) (err os.Error) {
	binary.BigEndian.PutUint64(ps.scratch[0:8], v)
	_, err = writer.Write(ps.scratch[0:8])
	return
}

// read/write float32.
func (ps *PacketSerializer) readFloat32(reader io.Reader) (v float32, err os.Error) {
	var vUint32 uint32
	if vUint32, err = ps.readUint32(reader); err != nil {
		return
	}
	return math.Float32frombits(vUint32), nil
}
func (ps *PacketSerializer) writeFloat32(writer io.Writer, v float32) (err os.Error) {
	return ps.writeUint32(writer, math.Float32bits(v))
}

// read/write float64.
func (ps *PacketSerializer) readFloat64(reader io.Reader) (v float64, err os.Error) {
	var vUint64 uint64
	if vUint64, err = ps.readUint64(reader); err != nil {
		return
	}
	return math.Float64frombits(vUint64), nil
}
func (ps *PacketSerializer) writeFloat64(writer io.Writer, v float64) (err os.Error) {
	return ps.writeUint64(writer, math.Float64bits(v))
}
