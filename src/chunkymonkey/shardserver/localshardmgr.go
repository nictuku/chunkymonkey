package shardserver

import (
	"log"
	"sync"

	"chunkymonkey/chunkstore"
	"chunkymonkey/entity"
	"chunkymonkey/gamerules"
	"chunkymonkey/stub"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

// localShardConnection implements IShardConnection for LocalShardManager.
type localShardConnection struct {
	entityId EntityId
	player   stub.IPlayerConnection
	shard    *ChunkShard
}

func newLocalShardConnection(entityId EntityId, player stub.IPlayerConnection, shard *ChunkShard) *localShardConnection {
	return &localShardConnection{
		entityId: entityId,
		player:   player,
		shard:    shard,
	}
}

func (conn *localShardConnection) Disconnect() {
	// TODO This inefficiently unsubscribes from all chunks, even if not
	// subscribed to.
	conn.shard.EnqueueAllChunks(func(chunk *Chunk) {
		chunk.reqUnsubscribeChunk(conn.entityId, false)
	})
}

func (conn *localShardConnection) Enqueue(fn func()) {
	conn.shard.Enqueue(fn)
}

func (conn *localShardConnection) ReqSubscribeChunk(chunkLoc ChunkXz) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqSubscribeChunk(conn.entityId, conn.player)
	})
}

func (conn *localShardConnection) ReqUnsubscribeChunk(chunkLoc ChunkXz) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqUnsubscribeChunk(conn.entityId, true)
	})
}

func (conn *localShardConnection) ReqMulticastPlayers(chunkLoc ChunkXz, exclude EntityId, packet []byte) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqMulticastPlayers(exclude, packet)
	})
}

func (conn *localShardConnection) ReqAddPlayerData(chunkLoc ChunkXz, name string, position AbsXyz, look LookDegrees, held ItemTypeId) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqAddPlayerData(conn.entityId, name, position, look, held)
	})
}

func (conn *localShardConnection) ReqRemovePlayerData(chunkLoc ChunkXz, isDisconnect bool) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqRemovePlayerData(conn.entityId, isDisconnect)
	})
}

func (conn *localShardConnection) ReqSetPlayerPositionLook(chunkLoc ChunkXz, position AbsXyz, look LookDegrees, moved bool) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqSetPlayerPositionLook(conn.entityId, position, look, moved)
	})
}

func (conn *localShardConnection) ReqHitBlock(held slot.Slot, target BlockXyz, digStatus DigStatus, face Face) {
	chunkLoc := target.ToChunkXz()
	log.Println("target", target)
	conn.shard.EnqueueOnChunk(*chunkLoc, func(chunk *Chunk) {
		chunk.reqHitBlock(conn.player, held, digStatus, &target, face)
	})
}

func (conn *localShardConnection) ReqInteractBlock(held slot.Slot, target BlockXyz, face Face) {
	chunkLoc := target.ToChunkXz()

	conn.shard.EnqueueOnChunk(*chunkLoc, func(chunk *Chunk) {
		chunk.reqInteractBlock(conn.player, held, &target, face)
	})
}

func (conn *localShardConnection) ReqPlaceItem(target BlockXyz, slot slot.Slot) {
	chunkLoc, _ := target.ToChunkLocal()

	conn.shard.EnqueueOnChunk(*chunkLoc, func(chunk *Chunk) {
		chunk.reqPlaceItem(conn.player, &target, &slot)
	})
}

func (conn *localShardConnection) ReqTakeItem(chunkLoc ChunkXz, entityId EntityId) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqTakeItem(conn.player, entityId)
	})
}

func (conn *localShardConnection) ReqDropItem(content slot.Slot, position AbsXyz, velocity AbsVelocity) {
	chunkLoc := position.ToChunkXz()
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqDropItem(conn.player, &content, &position, &velocity)
	})
}

// LocalShardManager contains all chunk shards and can look them up. It
// implements IShardConnecter and is for use in hosting all shards in the local
// process.
type LocalShardManager struct {
	entityMgr  *entity.EntityManager
	chunkStore chunkstore.IChunkStore
	gameRules  *gamerules.GameRules
	shards     map[uint64]*ChunkShard
	lock       sync.Mutex
}

func NewLocalShardManager(chunkStore chunkstore.IChunkStore, entityMgr *entity.EntityManager, gameRules *gamerules.GameRules) *LocalShardManager {
	return &LocalShardManager{
		entityMgr:  entityMgr,
		chunkStore: chunkStore,
		gameRules:  gameRules,
		shards:     make(map[uint64]*ChunkShard),
	}
}

func (mgr *LocalShardManager) getShard(loc ShardXz) *ChunkShard {
	shardKey := loc.Key()
	if shard, ok := mgr.shards[shardKey]; ok {
		// Shard already exists.
		return shard
	}

	// Create shard.
	shard := NewChunkShard(mgr, loc)
	mgr.shards[shardKey] = shard
	go shard.serve()

	return shard
}

func (mgr *LocalShardManager) ShardConnect(entityId EntityId, player stub.IPlayerConnection, shardLoc ShardXz) stub.IShardConnection {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	shard := mgr.getShard(shardLoc)
	return newLocalShardConnection(entityId, player, shard)
}

// TODO remove Enqueue* methods

// EnqueueAllChunks runs a given function on all loaded chunks.
func (mgr *LocalShardManager) EnqueueAllChunks(fn func(chunk *Chunk)) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	for _, shard := range mgr.shards {
		shard.EnqueueAllChunks(fn)
	}
}

// EnqueueOnChunk runs a function on the chunk at the given location. If the
// chunk does not exist, it does nothing.
func (mgr *LocalShardManager) EnqueueOnChunk(loc ChunkXz, fn func(chunk *Chunk)) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	shard := mgr.getShard(loc.ToShardXz())
	shard.EnqueueOnChunk(loc, fn)
}
