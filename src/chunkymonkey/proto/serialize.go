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
		if _, err = io.ReadFull(reader, ps.scratch[0:1]); err != nil {
			return
		}
		value.SetBool(ps.scratch[0] != 0)

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

		// Floating point types:

	case reflect.Float32:
		if _, err = io.ReadFull(reader, ps.scratch[0:4]); err != nil {
			return
		}
		value.SetFloat(float64(math.Float32frombits(binary.BigEndian.Uint32(ps.scratch[0:4]))))

	case reflect.Float64:
		if _, err = io.ReadFull(reader, ps.scratch[0:8]); err != nil {
			return
		}
		value.SetFloat(math.Float64frombits(binary.BigEndian.Uint64(ps.scratch[0:8])))

	case reflect.String:
		// TODO Maybe the tag field could/should suggest a max length.
		if _, err = io.ReadFull(reader, ps.scratch[0:2]); err != nil {
			return
		}
		length := int16(binary.BigEndian.Uint16(ps.scratch[0:2]))
		if length < 0 {
			return ErrorLengthNegative
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
		if value.Bool() {
			ps.scratch[0] = 1
		} else {
			ps.scratch[0] = 0
		}
		_, err = writer.Write(ps.scratch[0:1])

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

		// Floating point types:

	case reflect.Float32:
		binary.BigEndian.PutUint32(ps.scratch[0:4], math.Float32bits(float32(value.Float())))
		_, err = writer.Write(ps.scratch[0:4])
	case reflect.Float64:
		binary.BigEndian.PutUint64(ps.scratch[0:8], math.Float64bits(value.Float()))
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
