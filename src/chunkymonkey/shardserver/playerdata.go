package shardserver

import (
	"io"
	"os"

	"chunkymonkey/item"
	"chunkymonkey/proto"
	. "chunkymonkey/types"
)

const (
	// Assumed values for size of player axis-aligned bounding box (AAB).
	playerAabH = AbsCoord(0.75) // Each side of player.
	playerAabY = AbsCoord(2.00) // From player's feet position upwards.
)

// playerData represents a Chunk's knowledge about a player. Only one Chunk has
// this data at a time. This data is occasionally updated from the frontend
// server.
type playerData struct {
	entityId   EntityId
	name       string
	position   AbsXyz
	look       LookDegrees
	heldItemId ItemTypeId
	// TODO Armor data.
}

func (player *playerData) sendSpawn(writer io.Writer) os.Error {
	return proto.WriteNamedEntitySpawn(
		writer,
		player.entityId, player.name,
		player.position.ToAbsIntXyz(),
		player.look.ToLookBytes(),
		player.heldItemId,
	)
	// TODO Armor packet(s).
}

func (player *playerData) sendPositionLook(writer io.Writer) os.Error {
	return proto.WriteEntityTeleport(
		writer,
		player.entityId,
		player.position.ToAbsIntXyz(),
		player.look.ToLookBytes())
}

func (player *playerData) OverlapsItem(item *item.Item) bool {
	// TODO note that calling this function repeatedly is not as efficient as it
	// could be.

	minX := player.position.X - playerAabH
	maxX := player.position.X + playerAabH
	minZ := player.position.Z - playerAabH
	maxZ := player.position.Z + playerAabH
	minY := player.position.Y
	maxY := player.position.Y + playerAabY

	pos := item.Position()

	return pos.X >= minX && pos.X <= maxX && pos.Y >= minY && pos.Y <= maxY && pos.Z >= minZ && pos.Z <= maxZ
}
