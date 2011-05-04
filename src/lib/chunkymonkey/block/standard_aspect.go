package block

import (
	"log"

	"chunkymonkey/item"
	. "chunkymonkey/types"
)

type BlockDropItem struct {
	DroppedItem ItemTypeId
	Probability byte // Probabilities specified as a percentage
	Count       ItemCount
	CopyData    bool
}

func makeStandardAspect() (aspect IBlockAspect) {
	return &StandardAspect{}
}

// Behaviour of a "standard" block. A StandardAspect block is one that is
// diggable, and drops items in a simple manner. StandardAspect blocks do not
// use block metadata.
type StandardAspect struct {
	// Items, up to one of which will potentially spawn when block destroyed.
	DroppedItems []BlockDropItem
	BreakOn      DigStatus
}

func (aspect *StandardAspect) Name() string {
	return "Standard"
}

func (aspect *StandardAspect) Hit(chunk IChunkBlock, blockLoc *BlockXyz, blockData byte, digStatus DigStatus) (destroyed bool) {
	if aspect.BreakOn != digStatus {
		return
	}

	destroyed = true

	if len(aspect.DroppedItems) > 0 {
		rand := chunk.GetRand()
		// Possibly drop item(s)
		r := byte(rand.Intn(100))
		for _, dropItem := range aspect.DroppedItems {
			if dropItem.Probability > r {
				itemType, ok := chunk.GetItemType(dropItem.DroppedItem)

				if !ok {
					log.Printf(
						"Warning: tried to create item with type ID #%d - "+
							"but no such item type is defined. block and item "+
							"definitions out of sync?",dropItem.DroppedItem)

					break
				}

				var itemData ItemData
				if dropItem.CopyData {
					itemData = ItemData(blockData)
				} else {
					itemData = 0
				}

				for i := dropItem.Count; i > 0; i-- {
					position := blockLoc.ToAbsXyz()
					position.X += AbsCoord(blockItemSpawnFromEdge + rand.Float64()*(1-2*blockItemSpawnFromEdge))
					position.Y += AbsCoord(blockItemSpawnFromEdge)
					position.Z += AbsCoord(blockItemSpawnFromEdge + rand.Float64()*(1-2*blockItemSpawnFromEdge))
					chunk.AddItem(
						item.NewItem(
							itemType, 1, itemData,
							position,
							&AbsVelocity{0, 0, 0}))
				}
				break
			}
			r -= dropItem.Probability
		}
	}

	return
}

func (aspect *StandardAspect) Interact(player IBlockPlayer) {
}
