package types

import (
	"math"
)

// Defines the basic types such as ID types, and world units.

type TimeOfDay int64

const (
	DayTicksPerDay    = TimeOfDay(24000)
	DayTicksPerSecond = TimeOfDay(20)
)

// 1 "TickTime" is the duration of a server "tick". This value is intended for
// use in sub-tick physics calculations.
type TickTime float64

const (
	TicksPerSecond      = 5
	DayTicksPerTick     = DayTicksPerSecond / TicksPerSecond
	NanosecondsInSecond = 1e9
)

type RandomSeed int64

// Which 'world'?
type DimensionId int8

const (
	DimensionNether = DimensionId(-1)
	DimensionNormal = DimensionId(0)
)

// Item-related types

// Item type ID
type ItemTypeId int16

const ItemTypeIdNull = ItemTypeId(-1)

// Item metadata. The meaning of this varies depending upon the item type. In
// the case of tools/armor it indicates "uses" or "damage".
type ItemData int16

// How many items are in a stack/slot etc.
type ItemCount int8

// Entity-related types

type EntityId int32

// The type of mob
type EntityMobType byte

const (
	MobTypeIdCreeper      = EntityMobType(50)
	MobTypeIdSkeleton     = EntityMobType(51)
	MobTypeIdSpider       = EntityMobType(52)
	MobTypeIdGiantZombie  = EntityMobType(53)
	MobTypeIdZombie       = EntityMobType(54)
	MobTypeIdSlime        = EntityMobType(55)
	MobTypeIdGhast        = EntityMobType(56)
	MobTypeIdZombiePigman = EntityMobType(57)
	MobTypeIdPig          = EntityMobType(90)
	MobTypeIdSheep        = EntityMobType(91)
	MobTypeIdCow          = EntityMobType(92)
	MobTypeIdHen          = EntityMobType(93)
	MobTypeIdSquid        = EntityMobType(94)
	MobTypeIdWolf         = EntityMobType(95)
)

type EntityStatus byte

type EntityAnimation byte

const (
	EntityAnimationNone     = EntityAnimation(0)
	EntityAnimationSwingArm = EntityAnimation(1)
	EntityAnimationDamage   = EntityAnimation(2)
	EntityAnimationUnknown1 = EntityAnimation(102)
	EntityAnimationCrouch   = EntityAnimation(104)
	EntityAnimationUncrouch = EntityAnimation(105)
)

type EntityAction byte

const (
	EntityActionCrouch   = EntityAction(1)
	EntityActionUncrouch = EntityAction(2)
)

type ObjTypeId int8

const (
	ObjTypeIdBoat           = ObjTypeId(1)
	ObjTypeIdMinecart       = ObjTypeId(10)
	ObjTypeIdStorageCart    = ObjTypeId(11)
	ObjTypeIdPoweredCart    = ObjTypeId(12)
	ObjTypeIdActivatedTnt   = ObjTypeId(50)
	ObjTypeIdArrow          = ObjTypeId(60)
	ObjTypeIdThrownSnowball = ObjTypeId(61)
	ObjTypeIdThrownEgg      = ObjTypeId(62)
	ObjTypeIdFallingSand    = ObjTypeId(70)
	ObjTypeIdFallingGravel  = ObjTypeId(71)
	ObjTypeIdFishingFloat   = ObjTypeId(90)
)

type PaintingTypeId int32

type InstrumentId byte

const (
	InstrumentIdDoubleBass = InstrumentId(1)
	InstrumentIdSnareDrum  = InstrumentId(2)
	InstrumentIdSticks     = InstrumentId(3)
	InstrumentIdBassDrum   = InstrumentId(4)
	InstrumentIdHarp       = InstrumentId(5)
)

type NotePitch byte

const (
	NotePitchMin = NotePitch(0)
	NotePitchMax = NotePitch(24)
)

// Block-related types

type BlockId byte

const (
	BlockIdMin = 0
	BlockIdAir = BlockId(0)
	BlockIdMax = 255
)

// Block face (0-5)
type Face int8

// Used when a block face is not appropriate to the situation, but block
// location data passed (such as using an item not on a block).
const (
	FaceNull     = Face(-1)
	FaceMinValid = 0
	FaceBottom   = 0
	FaceTop      = 1
	FaceWest     = 2
	FaceEast     = 3
	FaceNorth    = 4
	FaceSouth    = 5
	FaceMaxValid = 5
)

func (f Face) GetDxyz() (dx BlockCoord, dy BlockYCoord, dz BlockCoord) {
	switch f {
	case FaceBottom:
		dy = -1
	case FaceTop:
		dy = 1
	case FaceWest:
		dz = -1
	case FaceEast:
		dz = 1
	case FaceNorth:
		dx = -1
	case FaceSouth:
		dx = 1
	}
	return
}

// Action-related types and constants

type DigStatus byte

const (
	DigStarted    = DigStatus(0)
	DigBlockBroke = DigStatus(2)
	// TODO Investigate what this value means:
	DigDropItem = DigStatus(4)
)

// Window/inventory-related types and constants

// ID specifying which slotted window, such as inventory
type WindowId int8

const (
	WindowIdCursor    = WindowId(-1)
	WindowIdInventory = WindowId(0)
)

type InvTypeId int8

const (
	InvTypeIdChest     = InvTypeId(0)
	InvTypeIdWorkbench = InvTypeId(1)
	InvTypeIdFurnace   = InvTypeId(2)
	InvTypeIdDispenser = InvTypeId(3)
)

// ID of the slow in inventory or other item-slotted window element
type SlotId int16

const (
	SlotIdCursor = SlotId(-1)
	SlotIdNull   = SlotId(999) // Clicked outside window.
)

type PrgBarId int16

const (
	PrgBarIdFurnaceProgress = PrgBarId(0)
	PrgBarIdFurnaceFire     = PrgBarId(1)
)

type PrgBarValue int16

// ID specifying a player statistic.
type StatisticId int32

// Transaction ID
type TxId int16

// Movement-related types and constants

// VelocityComponent in millipixels / tick
type VelocityComponent int16

const (
	VelocityComponentMax = 28800
	VelocityComponentMin = -28800

	MaxVelocityBlocksPerTick = VelocityComponentMax / AbsVelocityCoord(MilliPixelsPerBlock)
	MinVelocityBlocksPerTick = VelocityComponentMin / AbsVelocityCoord(MilliPixelsPerBlock)
)

type Velocity struct {
	X, Y, Z VelocityComponent
}

type AbsVelocityCoord AbsCoord

func (v AbsVelocityCoord) ToVelocityComponent() VelocityComponent {
	return VelocityComponent(v * MilliPixelsPerBlock)
}

type AbsVelocity struct {
	X, Y, Z AbsVelocityCoord
}

func (v *AbsVelocity) ToVelocity() *Velocity {
	return &Velocity{
		v.X.ToVelocityComponent(),
		v.Y.ToVelocityComponent(),
		v.Z.ToVelocityComponent(),
	}
}

func (v *AbsVelocityCoord) Constrain() {
	if *v > MaxVelocityBlocksPerTick {
		*v = MaxVelocityBlocksPerTick
	} else if *v < MinVelocityBlocksPerTick {
		*v = MinVelocityBlocksPerTick
	}
}

// Relative movement, using same units as AbsIntCoord, but in byte form so
// constrained
type RelMoveCoord int8

type RelMove struct {
	X, Y, Z RelMoveCoord
}

// Angle-related types and constants

const (
	DegreesToBytes = 256.0 / 360.0
)

// An angle, where there are 256 units in a circle.
type AngleBytes byte

// An angle in degrees
type AngleDegrees float32

func (d *AngleDegrees) ToAngleBytes() AngleBytes {
	norm := math.Fmod(float64(*d), 360)
	if norm < 0 {
		norm = 360 + norm
	}
	return AngleBytes(norm * DegreesToBytes)
}

type LookDegrees struct {
	// Pitch is -ve when looking above the horizontal, and +ve below
	Yaw, Pitch AngleDegrees
}

func (l *LookDegrees) ToLookBytes() *LookBytes {
	return &LookBytes{
		l.Yaw.ToAngleBytes(),
		l.Pitch.ToAngleBytes(),
	}
}

type LookBytes struct {
	Yaw, Pitch AngleBytes
}

type OrientationDegrees struct {
	Yaw, Pitch, Roll AngleDegrees
}

type OrientationBytes struct {
	Yaw, Pitch, Roll AngleBytes
}

// Cardinal directions
type ChunkSideDir int

// Note that these constants are different to those for Face. These index
// directly into arrays of fixed length.
const (
	ChunkSideEast  = 0
	ChunkSideSouth = 1
	ChunkSideWest  = 2
	ChunkSideNorth = 3
)

func (d ChunkSideDir) GetDxz() (dx, dz ChunkCoord) {
	switch d {
	case ChunkSideEast:
		dz = 1
	case ChunkSideSouth:
		dx = 1
	case ChunkSideWest:
		dz = -1
	case ChunkSideNorth:
		dx = -1
	}
	return
}

func (d ChunkSideDir) GetOpposite() ChunkSideDir {
	switch d {
	case ChunkSideEast:
		return ChunkSideWest
	case ChunkSideSouth:
		return ChunkSideNorth
	case ChunkSideWest:
		return ChunkSideEast
	case ChunkSideNorth:
		return ChunkSideSouth
	}
	// Should not happen (should we panic on this?)
	return ChunkSideNorth
}

// Returns the direction that (dx,dz) is in. Exactly one of dx and dz must be
// -1 or 1, and the other must be 0, otherwide ok will return as false.
func DxzToDir(dx, dz int32) (dir ChunkSideDir, ok bool) {
	ok = true
	if dz == 0 {
		if dx == -1 {
			dir = ChunkSideNorth
		} else if dx == 1 {
			dir = ChunkSideSouth
		} else {
			ok = false
		}
	} else if dx == 0 {
		if dz == -1 {
			dir = ChunkSideWest
		} else if dz == 1 {
			dir = ChunkSideEast
		} else {
			ok = false
		}
	} else {
		ok = false
	}
	return
}

// Location-related types and constants

const (
	// Chunk coordinates can be converted to block coordinates
	ChunkSizeH = 16
	ChunkSizeY = 128

	// The area within which a client receives updates.
	ChunkRadius = 10
	// The radius in which all chunks must be sent before completing a client's
	// login process.
	MinChunkRadius = 2

	// Sometimes it is useful to convert block coordinates to pixels
	PixelsPerBlock = 32

	// Millipixels are used in velocity values
	MilliPixelsPerPixel = 1000
	MilliPixelsPerBlock = PixelsPerBlock * MilliPixelsPerPixel
)

// Specifies exact world distance in blocks (floating point)
type AbsCoord float64

type AbsXyz struct {
	X, Y, Z AbsCoord
}

func (p *AbsXyz) Copy() *AbsXyz {
	return &AbsXyz{p.X, p.Y, p.Z}
}

// Updates a ChunkXz with the chunk position that the AbsXyz is within.
func (p *AbsXyz) UpdateChunkXz(chunkLoc *ChunkXz) {
	chunkLoc.X = ChunkCoord(math.Floor(float64(p.X / ChunkSizeH)))
	chunkLoc.Z = ChunkCoord(math.Floor(float64(p.Z / ChunkSizeH)))
}

func (p *AbsXyz) ApplyVelocity(dt TickTime, v *AbsVelocity) {
	p.X += AbsCoord(float64(v.X) * float64(dt))
	p.Y += AbsCoord(float64(v.Y) * float64(dt))
	p.Z += AbsCoord(float64(v.Z) * float64(dt))
}

func (p *AbsXyz) ToAbsIntXyz() *AbsIntXyz {
	return &AbsIntXyz{
		AbsIntCoord(p.X * PixelsPerBlock),
		AbsIntCoord(p.Y * PixelsPerBlock),
		AbsIntCoord(p.Z * PixelsPerBlock),
	}
}

func (p *AbsXyz) ToBlockXyz() *BlockXyz {
	return &BlockXyz{
		BlockCoord(math.Floor(float64(p.X))),
		BlockYCoord(math.Floor(float64(p.Y))),
		BlockCoord(math.Floor(float64(p.Z))),
	}
}

// Specifies approximate world distance in pixels (absolute / PixelsPerBlock)
type AbsIntCoord int32

type AbsIntXyz struct {
	X, Y, Z AbsIntCoord
}

func (p *AbsIntXyz) ToBlockXyz() *BlockXyz {
	return &BlockXyz{
		BlockCoord(p.X / PixelsPerBlock),
		BlockYCoord(p.Y / PixelsPerBlock),
		BlockCoord(p.Z / PixelsPerBlock),
	}
}

// Coordinate of a chunk in the world (block / 16)
type ChunkCoord int32

func (c ChunkCoord) Abs() ChunkCoord {
	if c < 0 {
		return -c
	}
	return c
}

type ChunkXz struct {
	X, Z ChunkCoord
}

// Returns the world BlockXyz position of the (0, 0, 0) block in the chunk
func (chunkLoc *ChunkXz) GetChunkCornerBlockXY() *BlockXyz {
	return &BlockXyz{
		BlockCoord(chunkLoc.X) * ChunkSizeH,
		0,
		BlockCoord(chunkLoc.Z) * ChunkSizeH,
	}
}

// Convert a position within a chunk to a block position within the world
func (chunkLoc *ChunkXz) ToBlockXyz(subLoc *SubChunkXyz) *BlockXyz {
	return &BlockXyz{
		BlockCoord(chunkLoc.X)*ChunkSizeH + BlockCoord(subLoc.X),
		BlockYCoord(subLoc.Y),
		BlockCoord(chunkLoc.Z)*ChunkSizeH + BlockCoord(subLoc.Z),
	}
}

// Converts a chunk location into a key suitable for using in a hash.
func (chunkLoc *ChunkXz) ChunkKey() uint64 {
	return uint64(chunkLoc.X)<<32 | uint64(uint32(chunkLoc.Z))
}

// Size of a sub-chunk
type SubChunkSizeCoord byte

// Coordinate of a block within a chunk
type SubChunkCoord byte

type SubChunkSize struct {
	X, Y, Z SubChunkSizeCoord
}

type SubChunkXyz struct {
	X, Y, Z SubChunkCoord
}

// Coordinate of a block within the world
type BlockCoord int32
type BlockYCoord int8

type BlockXyz struct {
	X BlockCoord
	Y BlockYCoord
	Z BlockCoord
}

// Test if a block location is not appropriate to the situation, but block
// location data passed (such as using an item not on a block).
func (b *BlockXyz) IsNull() bool {
	return b.Y == -1 && b.X == -1 && b.Z == -1
}

// Convert an (x, z) absolute coordinate pair to chunk coordinates
func (abs *AbsXyz) ToChunkXz() (chunkXz *ChunkXz) {
	return &ChunkXz{
		ChunkCoord(abs.X / ChunkSizeH),
		ChunkCoord(abs.Z / ChunkSizeH),
	}
}

// Convert (x, z) absolute integer coordinates to chunk coordinates
func (abs *AbsIntXyz) ToChunkXz() *ChunkXz {
	chunkX, _ := coordDivMod(int32(abs.X), ChunkSizeH*PixelsPerBlock)
	chunkZ, _ := coordDivMod(int32(abs.Z), ChunkSizeH*PixelsPerBlock)

	return &ChunkXz{
		ChunkCoord(chunkX),
		ChunkCoord(chunkZ),
	}
}

func (abs *AbsIntXyz) IAdd(dx, dy, dz AbsIntCoord) {
	abs.X += dx
	abs.Y += dy
	abs.Z += dz
}

func coordDivMod(num, denom int32) (div, mod int32) {
	div = num / denom
	mod = num % denom
	if mod < 0 {
		mod += denom
		div -= 1
	}
	return
}

// Convert an (x, z) block coordinate pair to chunk coordinates and the
// coordinates of the block within the chunk
func (blockLoc *BlockXyz) ToChunkLocal() (chunkLoc *ChunkXz, subLoc *SubChunkXyz) {
	chunkX, subX := coordDivMod(int32(blockLoc.X), ChunkSizeH)
	chunkZ, subZ := coordDivMod(int32(blockLoc.Z), ChunkSizeH)

	chunkLoc = &ChunkXz{ChunkCoord(chunkX), ChunkCoord(chunkZ)}
	subLoc = &SubChunkXyz{SubChunkCoord(subX), SubChunkCoord(blockLoc.Y), SubChunkCoord(subZ)}
	return
}

func (blockLoc *BlockXyz) ToAbsIntXyz() *AbsIntXyz {
	return &AbsIntXyz{
		AbsIntCoord(blockLoc.X) * PixelsPerBlock,
		AbsIntCoord(blockLoc.Y) * PixelsPerBlock,
		AbsIntCoord(blockLoc.Z) * PixelsPerBlock,
	}
}

func (blockLoc *BlockXyz) ToAbsXyz() *AbsXyz {
	return &AbsXyz{
		AbsCoord(blockLoc.X),
		AbsCoord(blockLoc.Y),
		AbsCoord(blockLoc.Z),
	}
}

// Misc. types and constants

type ChunkLoadMode byte

const (
	// Client should unload the chunk
	ChunkUnload = ChunkLoadMode(0)

	// Client should initialise the chunk
	ChunkInit = ChunkLoadMode(1)
)
