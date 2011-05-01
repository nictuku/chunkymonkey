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
	expVarMobSpawnCount   *expvar.Int
)

func init() {
	expVarMobSpawnCount = expvar.NewInt("mob-spawn-count")
}

type Mob struct {
	Entity
	mobType  EntityMobType
	position AbsXyz
	look     LookDegrees
	metadata []proto.EntityMetadata
	lock     sync.Mutex
}

func NewMob(mobType *EntityMobType, position *AbsXyz) (mob *Mob) {
	mob = &Mob{
		mobType:  *mobType,
		position: *position,
		look:     LookDegrees{130, 0},
		metadata: []proto.EntityMetadata{
			proto.EntityMetadata{0, 17, byte(0)},
			proto.EntityMetadata{0, 0, byte(0)},
			proto.EntityMetadata{0, 16, byte(255)}},
	}
	log.Println("new mob", mob)
	return
}

func (mob *Mob) GetEntityId() EntityId {
	return mob.EntityId
}

func (mob *Mob) GetEntity() *Entity {
	return &mob.Entity
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
