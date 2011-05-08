package block

import (
	"log"

	"chunkymonkey/inventory"
	. "chunkymonkey/types"
)

func makeWorkbenchAspect() (aspect IBlockAspect) {
	return &WorkbenchAspect{}
}

// WorkbenchAspect is the behaviour for the workbench block that allows 3x3
// crafting.
type WorkbenchAspect struct {
	StandardAspect
}

func (aspect *WorkbenchAspect) Name() string {
	return "Workbench"
}

func (aspect *WorkbenchAspect) Interact(instance *BlockInstance, player IBlockPlayer) {
	workbenchInv, ok := instance.Chunk.GetBlockExtra(&instance.SubLoc).(*inventory.WorkbenchInventory)
	if !ok {
		// TODO have the inventory stop existing when all players unsubscribe from
		// it. This is merely to reclaim a little memory.
		ejectItems := func() { aspect.ejectItems(instance) }
		workbenchInv = inventory.NewWorkbenchInventory(ejectItems, instance.Chunk.GetRecipeSet())
		instance.Chunk.SetBlockExtra(&instance.SubLoc, workbenchInv)
	}
	player.OpenWindow(InvTypeIdWorkbench, workbenchInv)
}

func (aspect *WorkbenchAspect) ejectItems(instance *BlockInstance) {
	instance.Chunk.EnqueueGeneric(func(_ interface{}) {
		workbenchInv, ok := instance.Chunk.GetBlockExtra(&instance.SubLoc).(*inventory.WorkbenchInventory)
		if !ok {
			return
		}

		items := workbenchInv.TakeAllItems()
		for _, slot := range items {
			log.Printf("spawning from slot %#v", slot)
			spawnItemInBlock(instance, slot.ItemType, slot.Count, slot.Data)
		}
	})
}
