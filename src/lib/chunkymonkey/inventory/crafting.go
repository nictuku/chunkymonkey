package inventory

import (
	"chunkymonkey/gamerules"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

// Inventory with extended function to perform crafting. It assumes that slot 0
// is the output, and that the remaining slots are inputs.
type CraftingInventory struct {
	Inventory
	width, height int
	gameRules     *gamerules.GameRules
}

func (inv *CraftingInventory) Init(width, height int, gameRules *gamerules.GameRules) {
	inv.Inventory.Init(1 + width*height)
	inv.width = width
	inv.height = height
	inv.gameRules = gameRules
}

// Click handles window clicks from a user with special handling for crafting.
func (inv *CraftingInventory) Click(slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool) (accepted bool) {
	if slotId == 0 {
		// Player may only *take* the *whole* stack from the output slot.
		accepted = inv.Inventory.TakeOnlyClick(
			slotId, cursor, rightClick, shiftClick)
	} else {
		// Player may interact with the input slots like any other slot.
		accepted = inv.Inventory.StandardClick(
			slotId, cursor, rightClick, shiftClick)
	}

	if !accepted {
		return
	}

	inv.lock.Lock()
	defer inv.lock.Unlock()

	if slotId == 0 {
		// Player took items from the output slot. Subtract 1 count from each
		// non-empty input slot.
		for i := 1; i < len(inv.slots); i++ {
			inv.slots[i].Decrement()
		}
	}

	// Match recipe and set output slot.
	inv.slots[0] = inv.gameRules.Recipes.Match(inv.width, inv.height, inv.slots[1:])
	inv.slotUpdate(&inv.slots[0], 0)

	return
}
