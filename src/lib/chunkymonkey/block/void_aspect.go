package block

import (
	. "chunkymonkey/types"
)

func makeVoidAspect() (aspect IBlockAspect) {
	return &VoidAspect{}
}

// Behaviour of a "void" block which has no behaviour.
type VoidAspect struct{}

func (aspect *VoidAspect) Name() string {
	return "Void"
}

func (aspect *VoidAspect) Hit(chunk IChunkBlock, blockLoc *BlockXyz, blockData byte, digStatus DigStatus) (destroyed bool) {
	destroyed = false
	return
}

func (aspect *VoidAspect) Interact(player IBlockPlayer) {
}
