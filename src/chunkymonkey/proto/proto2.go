package proto

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
	"log"
	"math"
	"os"
	"reflect"

	. "chunkymonkey/types"
)

// Possible error values for reading and writing packets.
var (
	ErrorPacketNotPtr      = os.NewError("packet not passed as a pointer")
	ErrorPacketNil         = os.NewError("packet was passed by a nil pointer")
	ErrorLengthNegative    = os.NewError("length was negative")
	ErrorStrTooLong        = os.NewError("string was too long")
	ErrorBadPacketData     = os.NewError("packet data well-formed but contains out of range values")
	ErrorBadChunkDataSize  = os.NewError("map chunk data length mismatches with size")
	ErrorMismatchingValues = os.NewError("packet data contains mismatching values")
	ErrorInternal          = os.NewError("implementation problem with packetization")
)

var (
	// Space to dump unwanted data into. As the contents of this aren't used, it
	// doesn't require syncronization.
	dump [4096]byte
)

// Packet definitions.

type PacketKeepAlive struct {
	Id int32
}

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

type PacketHandshake struct {
	UsernameOrHash string
}

type PacketChatMessage struct {
	Message string
}

type PacketTimeUpdate struct {
	Time Ticks
}

type PacketEntityEquipment struct {
	EntityId   EntityId
	Slot       SlotId
	ItemTypeId ItemTypeId
	Data       ItemData
}

type PacketSpawnPosition struct {
	X BlockCoord
	Y int32
	Z BlockCoord
}

type PacketUseEntity struct {
	User      EntityId
	Target    EntityId
	LeftClick bool
}

type PacketUpdateHealth struct {
	Health         Health
	Food           FoodUnits
	FoodSaturation float32
}

type PacketRespawn struct {
	Dimension   DimensionId
	Difficulty  GameDifficulty
	GameType    GameType
	WorldHeight int16
	MapSeed     RandomSeed
}

type PacketOnGround struct {
	OnGround bool
}

type PacketPlayerPosition struct {
	X, Y, Stance, Z AbsCoord
	OnGround        bool
}

type PacketPlayerLook struct {
	Look LookDegrees
}

type PacketPlayerPositionLook struct {
	X, Y1, Y2, Z AbsCoord
	Look         LookDegrees
	OnGround     bool
}

type PacketPlayerBlockHit struct {
	Status   DigStatus
	Position BlockXyz
	Face     Face
}

type PacketPlayerBlockInteract struct {
	Position BlockXyz
	Face     Face
	Tool     ItemSlot
}

type PacketPlayerHoldingChange struct {
	SlotId SlotId
}

type PacketPlayerUseBed struct {
	EntityId EntityId
	Flag     byte
	Position BlockXyz
}

type PacketEntityAnimation struct {
	EntityId  EntityId
	Animation EntityAnimation
}

type PacketEntityAction struct {
	EntityId EntityId
	Action   EntityAction
}

type PacketNamedEntitySpawn struct {
	EntityId    EntityId
	Username    string
	Position    AbsIntXyz
	Rotation    LookBytes
	CurrentItem ItemTypeId
}

type PacketItemSpawn struct {
	EntityId    EntityId
	ItemTypeId  ItemTypeId
	Count       ItemCount
	Data        ItemData
	Position    AbsIntXyz
	Orientation OrientationBytes
}

type PacketItemCollect struct {
	CollectedItem EntityId
	Collector     EntityId
}

type PacketObjectSpawn struct {
	EntityId EntityId
	ObjType  ObjTypeId
	Position AbsIntXyz
}

type PacketMobSpawn struct {
	EntityId EntityId
	MobType  EntityMobType
	Position AbsIntXyz
	Look     LookBytes
}

type PacketPaintingSpawn struct {
	EntityId EntityId
	Title    string
	Position AbsIntXyz
	SideFace SideFace
}

type PacketExperienceOrb struct {
	EntityId EntityId
	Position AbsIntXyz
	Count    int16
}

type PacketEntityVelocity struct {
	EntityId EntityId
	Velocity Velocity
}

type PacketEntityDestroy struct {
	EntityId EntityId
}

type PacketEntity struct {
	EntityId EntityId
}

type PacketEntityRelMove struct {
	EntityId EntityId
	Move     RelMove
}

type PacketEntityLook struct {
	EntityId EntityId
	Look     LookBytes
}

type PacketEntityLookAndRelMove struct {
	EntityId EntityId
	Move     RelMove
	Look     LookBytes
}

type PacketEntityTeleport struct {
	EntityId EntityId
	Position AbsIntXyz
	Look     LookBytes
}

type PacketEntityStatus struct {
	EntityId EntityId
	Status   EntityStatus
}

type PacketEntityAttach struct {
	EntityId  EntityId
	VehicleId EntityId
}

type PacketEntityMetadata struct {
	EntityId EntityId
	Metadata EntityMetadataTable
}

type PacketEntityEffect struct {
	EntityId EntityId
	Effect   EntityEffect
	Value    int8
	Duration int16
}

type PacketEntityRemoveEffect struct {
	EntityId EntityId
	Effect   EntityEffect
}

type PacketExperience struct {
	Experience      int8
	Level           int8
	TotalExperience int16
}

type PacketPreChunk struct {
	ChunkLoc ChunkXz
	Mode     ChunkLoadMode
}

type PacketMapChunk struct {
	Corner BlockXyz
	Data   ChunkData
}

type PacketMultiBlockChange struct {
	ChunkLoc ChunkXz
	Changes  MultiBlockChanges
}

// IMinecraftMarshaler is the interface by which packet fields (or even whole
// packets) can customize their serialization. It will only work for
// struct-based types currently, as a hacky method of optimizing which packet
// fields are checked for this property.
// TODO Check if it doesn't really take that long and if it's therefore a
// pointless optimization.
type IMarshaler interface {
	MinecraftUnmarshal(reader io.Reader, ps *PacketSerializer) (err os.Error)
	MinecraftMarshal(writer io.Writer, ps *PacketSerializer) (err os.Error)
}

// EntityMetadataTable implements IMarshaler.
type EntityMetadataTable struct {
	Items []EntityMetadata
}

func (emt *EntityMetadataTable) MinecraftUnmarshal(reader io.Reader, ps *PacketSerializer) (err os.Error) {
	emt.Items, err = readEntityMetadataField(reader)
	return
}

func (emt *EntityMetadataTable) MinecraftMarshal(writer io.Writer, ps *PacketSerializer) (err os.Error) {
	return writeEntityMetadataField(writer, emt.Items)
}

// ItemSlot implements IMarshaler.
type ItemSlot struct {
	ItemTypeId ItemTypeId
	Count      ItemCount
	Data       ItemData
}

func (is *ItemSlot) MinecraftUnmarshal(reader io.Reader, ps *PacketSerializer) (err os.Error) {
	if err = binary.Read(reader, binary.BigEndian, &is.ItemTypeId); err != nil {
		return
	}

	if is.ItemTypeId == -1 {
		is.Count = 0
		is.Data = 0
	} else {
		var data struct {
			Count ItemCount
			Data  ItemData
		}
		if err = binary.Read(reader, binary.BigEndian, &data); err != nil {
			return
		}

		is.Count = data.Count
		is.Data = data.Data
	}
	return
}

func (is *ItemSlot) MinecraftMarshal(writer io.Writer, ps *PacketSerializer) (err os.Error) {
	if is.ItemTypeId == -1 {
		if err = binary.Write(writer, binary.BigEndian, &is.ItemTypeId); err != nil {
			return
		}
	} else {
		if err = binary.Write(writer, binary.BigEndian, is); err != nil {
			return
		}
	}
	return
}

// ChunkData implements IMarshaler.
type ChunkData struct {
	Size ChunkDataSize
	Data []byte
}

// ChunkDataSize contains the dimensions of the data represented inside ChunkData.
type ChunkDataSize struct {
	X, Y, Z byte
}

func (cd *ChunkData) MinecraftUnmarshal(reader io.Reader, ps *PacketSerializer) (err os.Error) {
	if err = ps.readData(reader, reflect.Indirect(reflect.ValueOf(&cd.Size))); err != nil {
		return
	}

	var length int32
	if err = binary.Read(reader, binary.BigEndian, &length); err != nil {
		return
	}

	zReader, err := zlib.NewReader(&io.LimitedReader{reader, int64(length)})
	if err != nil {
		return
	}
	defer zReader.Close()

	numBlocks := (int(cd.Size.X) + 1) * (int(cd.Size.Y) + 1) * (int(cd.Size.Z) + 1)
	expectedNumDataBytes := numBlocks + 3*(numBlocks>>1)
	cd.Data = make([]byte, expectedNumDataBytes)
	if _, err = io.ReadFull(zReader, cd.Data); err != nil {
		return
	}

	// Check that we're at the end of the compressed data to be sure of being in
	// sync with packet stream..
	n, err := io.ReadFull(zReader, dump[:])
	if err == os.EOF {
		err = nil
		if n > 0 {
			log.Printf("Unexpected extra chunk data byte of %d bytes", n)
		}
	} else if err == nil {
		log.Printf("Unexpected extra chunk data byte of at least %d bytes - assuming bad packet stream", n)
		return ErrorBadPacketData
	} else {
		// Other error.
		return err
	}

	return nil
}

func (cd *ChunkData) MinecraftMarshal(writer io.Writer, ps *PacketSerializer) (err os.Error) {
	if err = ps.writeData(writer, reflect.Indirect(reflect.ValueOf(&cd.Size))); err != nil {
		return
	}

	numBlocks := (int(cd.Size.X) + 1) * (int(cd.Size.Y) + 1) * (int(cd.Size.Z) + 1)
	expectedNumDataBytes := numBlocks + 3*(numBlocks>>1)
	if len(cd.Data) != expectedNumDataBytes {
		return ErrorBadChunkDataSize
	}

	buf := bytes.NewBuffer(make([]byte, 0, 4096))

	zWriter, err := zlib.NewWriter(buf)
	if err != nil {
		return
	}
	_, err = zWriter.Write(cd.Data)
	zWriter.Close()
	if err != nil {
		return
	}

	compressedBytes := buf.Bytes()
	if err = binary.Write(writer, binary.BigEndian, int32(len(compressedBytes))); err != nil {
		return
	}

	_, err = writer.Write(compressedBytes)
	return
}

// MultiBlockChanges implements IMarshaler.
type MultiBlockChanges struct {
	// Coords are packed x,y,z block coordinates relative to a chunk origin. Note
	// that these differ from the value for BlockIndex, which supplies conversion
	// methods for this purpose.
	Coords    []int16
	TypeIds   []byte
	BlockData []byte
}

func (mbc *MultiBlockChanges) MinecraftUnmarshal(reader io.Reader, ps *PacketSerializer) (err os.Error) {
	var numBlocks int16

	if err = binary.Read(reader, binary.BigEndian, &numBlocks); err != nil {
		return
	}

	if numBlocks < 0 {
		return ErrorLengthNegative
	} else if numBlocks == 0 {
		// Odd case.
		return nil
	}

	mbc.Coords = make([]int16, numBlocks)
	if err = binary.Read(reader, binary.BigEndian, mbc.Coords); err != nil {
		return
	}

	mbc.TypeIds = make([]byte, numBlocks)
	if _, err = io.ReadFull(reader, mbc.TypeIds); err != nil {
		return
	}

	mbc.BlockData = make([]byte, numBlocks)
	_, err = io.ReadFull(reader, mbc.BlockData)

	return
}

func (mbc *MultiBlockChanges) MinecraftMarshal(writer io.Writer, ps *PacketSerializer) (err os.Error) {
	numBlocks := len(mbc.Coords)
	if numBlocks != len(mbc.TypeIds) || numBlocks != len(mbc.BlockData) {
		return ErrorMismatchingValues
	}

	if err = binary.Write(writer, binary.BigEndian, int16(numBlocks)); err != nil {
		return
	}

	if err = binary.Write(writer, binary.BigEndian, mbc.Coords); err != nil {
		return
	}

	if _, err = writer.Write(mbc.TypeIds); err != nil {
		return
	}

	_, err = writer.Write(mbc.BlockData)
	return
}

// PacketSerializer reads and writes packets. It is not safe to use one
// simultaneously between multiple goroutines.
//
// It does not take responsibility for reading/writing the packet ID byte
// header. TODO Should it?
//
// It is designed to read and write struct types, and can only handle a few
// types - it is not a generalized serialization mechanism and isn't intended
// to be one. It exercises the freedom of having only limited types of packet
// structure partly for simplicity, and partly to allow for optimizations.
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
