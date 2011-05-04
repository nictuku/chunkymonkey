package mob

import (
	. "chunkymonkey/types"
)

type MobType struct {
	Id   EntityMobType
	Name string
}

type MobTypeMap map[EntityMobType]*MobType

var Mobs = MobTypeMap{
	MobTypeIdCreeper:      &Creeper,
	MobTypeIdSkeleton:     &Skeleton,
	MobTypeIdSpider:       &Spider,
	MobTypeIdGiantZombie:  &GiantZombie,
	MobTypeIdZombie:       &Zombie,
	MobTypeIdSlime:        &Slime,
	MobTypeIdGhast:        &Ghast,
	MobTypeIdZombiePigman: &ZombiePigman,
	MobTypeIdPig:          &Pig,
	MobTypeIdSheep:        &Sheep,
	MobTypeIdCow:          &Cow,
	MobTypeIdHen:          &Hen,
	MobTypeIdSquid:        &Squid,
	MobTypeIdWolf:         &Wolf,
}

var Creeper = MobType{MobTypeIdCreeper, "creeper"}
var Skeleton = MobType{MobTypeIdSkeleton, "skeleton"}
var Spider = MobType{MobTypeIdSpider, "spider"}
var GiantZombie = MobType{MobTypeIdGiantZombie, "giantzombie"}
var Zombie = MobType{MobTypeIdZombie, "zombie"}
var Slime = MobType{MobTypeIdSlime, "slime"}
var Ghast = MobType{MobTypeIdGhast, "ghast"}
var ZombiePigman = MobType{MobTypeIdZombiePigman, "zombiepigman"}
var Pig = MobType{MobTypeIdPig, "pig"}
var Sheep = MobType{MobTypeIdSheep, "sheep"}
var Cow = MobType{MobTypeIdCow, "cow"}
var Hen = MobType{MobTypeIdHen, "hen"}
var Squid = MobType{MobTypeIdSquid, "squid"}
var Wolf = MobType{MobTypeIdWolf, "wolf"}
