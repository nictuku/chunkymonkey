package mob

import (
	. "chunkymonkey/types"
)

type MobType struct {
	Id   EntityMobType
	Name string
}

type MobTypeMap map[EntityMobType]*MobType

// TODO: Load from JSON file instead.
var Creeper = MobType{MobTypeIdCreeper, "creeper"}
var Mobs = MobTypeMap{MobTypeIdCreeper: &Creeper}
