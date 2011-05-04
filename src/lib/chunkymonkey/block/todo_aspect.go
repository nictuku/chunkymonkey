package block

import (
	. "chunkymonkey/types"
)

func makeTodoAspect() (aspect IBlockAspect) {
	return &TodoAspect{}
}

// TodoAspect has the same behaviour as that of a "void" block -
// i.e none. However, its purpose is intended to mark a block type
// whose behaviour is still to be implemented. A comment allows for
// notes to be made, but provides no functional change.
type TodoAspect struct {
	Comment string
}

func (aspect *TodoAspect) Name() string {
	return "Todo"
}

func (aspect *TodoAspect) Hit(chunk IChunkBlock, blockLoc *BlockXyz, blockData byte, digStatus DigStatus) (destroyed bool) {
	destroyed = false
	return
}

func (aspect *TodoAspect) Interact(player IBlockPlayer) {
}
