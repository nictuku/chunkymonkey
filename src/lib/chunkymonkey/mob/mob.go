// This is a prototype that will be thrown away and rewritten after I find a
// usable design. Note, for example, the absense of locks.
package mob

import (
	"expvar"
	"log"
	"io"
	"os"
	"sync"

	. "chunkymonkey/entity"
	"chunkymonkey/proto"
	. "chunkymonkey/types"
)

var (
	expVarMobSpawnCount *expvar.Int
)

func init() {
	expVarMobSpawnCount = expvar.NewInt("mob-spawn-count")
}

// TODO: Add a Mob interface.
//var ActiveMobs = []*Mob{}

type Mob struct {
	Entity
	mobType  EntityMobType
	position AbsXyz
	look     LookDegrees
	metadata map[byte]byte
	lock     sync.Mutex
}

func (mob *Mob) SetPosition(pos AbsXyz) {
	mob.position = pos
}

func (mob *Mob) SetLook(look LookDegrees) {
	mob.look = look
}

func (mob *Mob) GetEntityId() EntityId {
	return mob.EntityId
}

func (mob *Mob) GetEntity() *Entity {
	return &mob.Entity
}

func (mob *Mob) SetBurning() {
	mob.metadata[0] = 0x01
}

func (mob *Mob) FormatMetadata() []proto.EntityMetadata {
	x := []proto.EntityMetadata{}
	for k, v := range mob.metadata {
		x = append(x, proto.EntityMetadata{0, k, v})
	}
	return x
}

func (mob *Mob) SendSpawn(writer io.Writer) (err os.Error) {
	err = proto.WriteEntitySpawn(
		writer,
		mob.Entity.EntityId,
		mob.mobType,
		mob.position.ToAbsIntXyz(),
		mob.look.ToLookBytes(),
		mob.FormatMetadata())
	if err != nil {
		expVarMobSpawnCount.Add(1)
	}
	return
}


// ======================= CREEPER ======================

var (
	creeperNormal   = byte(0)
	creeperBlueAura = byte(1)
)

type Creeper struct {
	*Mob
}

func NewCreeper() (c *Creeper) {
	m := &Mob{}
	c = &Creeper{m}

	c.Mob.mobType = CreeperType.Id
	c.Mob.look = LookDegrees{200, 0}
	c.Mob.metadata = map[byte]byte{}
	log.Println("new ", Mobs[c.Mob.mobType].Name)
	//ActiveMobs = append(ActiveMobs, mob)
	return c
}

func (c *Creeper) SetNormalStatus() {
	c.Mob.metadata[17] = creeperNormal
}

func (c *Creeper) CreeperSetBlueAura() {
	c.Mob.metadata[17] = creeperBlueAura
}
