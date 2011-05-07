package intercept_parse

import (
	"log"
	"chunkymonkey/proto"
	. "chunkymonkey/types"
)

func NewClientMobLogger() *EntitySpawnLogger {
	return &EntitySpawnLogger{NewClientParser()}
}

func NewServerMobLogger() *EntitySpawnLogger {
	return &EntitySpawnLogger{NewServerParser()}
}


type EntitySpawnLogger struct {
	*MessageParser
}

func (p *EntitySpawnLogger) PacketEntitySpawn(entityId EntityId, mobType EntityMobType, position *AbsIntXyz, look *LookBytes, metadata []proto.EntityMetadata) {
	p.printf("fooooooooooooo")
	// TODO(nictuku): Do fun stuff here.
}

func (p *EntitySpawnLogger) printf(format string, v ...interface{}) {
	log.Println("my custom printf")
	log.Printf(p.LogPrefix+format, v...)
	return
}
