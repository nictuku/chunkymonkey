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
	metadata []proto.EntityMetadata
	lock     sync.Mutex
}

func (mob *Mob) SetPosition(pos AbsXyz) {
	mob.position = pos
}

func (mob *Mob) GetEntityId() EntityId {
	return mob.EntityId
}

func (mob *Mob) GetEntity() *Entity {
	return &mob.Entity
}

func (mob *Mob) SetBurning() {
	// Assumes creeper, otherwise will panic.
	mob.metadata[1] = proto.EntityMetadata{0, 0, byte(1)}
}

func (mob *Mob) SendSpawn(writer io.Writer) (err os.Error) {
	err = proto.WriteEntitySpawn(
		writer,
		mob.Entity.EntityId,
		mob.mobType,
		mob.position.ToAbsIntXyz(),
		mob.look.ToLookBytes(),
		mob.metadata)
	if err != nil {
		expVarMobSpawnCount.Add(1)
	}
	return
}


// ======================= CREEPER ======================

type CreeperStatus int // index 17

const (
	creeperNormal = iota
	creeperBlueAura
)

type Creeper struct {
	*Mob
	Status CreeperStatus
}

func NewCreeper() (c *Creeper) {
	m := &Mob{}
	c = &Creeper{m, creeperNormal}

	log.Printf("%+v", CreeperType)
	c.Mob.mobType =  CreeperType.Id
	c.Mob.look = LookDegrees{200, 0}
	// I'm still unsure if this raw data should be kept here, or if we
	// should just have fields that formats the metadata wire data as
	// needed.
	c.Mob.metadata = []proto.EntityMetadata{
			proto.EntityMetadata{0, 17, byte(0)},
			proto.EntityMetadata{0, 0, byte(0)},
			proto.EntityMetadata{0, 16, byte(255)}}
	log.Println("new ", Mobs[c.Mob.mobType].Name)
	//ActiveMobs = append(ActiveMobs, mob)
	return c
}



// .. if the writer is set, sends an EntityMetadata packet with the change.
func (c *Creeper) CreeperSetBlueAura(writer io.Writer) {
	x := proto.EntityMetadata{0, 17, byte(1)}
	c.metadata[0] = x
	if writer != nil {
		// Send an EntityMetadata packet with the change.
	}
}
