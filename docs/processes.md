Processes
=========

This document intends to briefly lay out the processes that take place on a
Minecraft server. Note that "processes" in this sense does not refer to an OS
process or thread or goroutine or other implementation. It rather refers to a
game logical concept.

Note that while the 'Requires' notes seem to imply the use of read/write locks,
this is not the intent. Instead it is intended to indicate purely what
processes can run in parallel if they are to share memory on things. (E.g item
physics only requires read access to chunk data, so in theory a lot of items'
physics can be farmed out to multiple goroutines with no conflict, as long as
some controlling process pulls their results together before letting anything
alter block data).

Bear in mind while reading this that it might well be missing information or
otherwise incorrect. Corrections are welcome.


Block automata
--------------

The process by which blocks like fluids (water/lava), falling blocks
(sand/gravel), mechanisms (redstone, furnaces) change the block IDs, block data
etc. within a chunk.

Requires:
*   R/W to block data.


Item physics
------------

Notably items for pickup in Minecraft don't interact with each other or
anything else (apart from blocks) in any way. Their movement is only affected
by gravity, friction and blocks (solid or fluid).

Requires:
*   R to block data.
*   R/W to individual item data.


Mob behaviour
-------------

Mobs can be potentially quite fiddly. They need to have access to information
about players nearby (and potentially other mobs - wolves will certainly need
this when tamed and hunting for players). They also need some kind of "line of
sight" to check if they can see a player or other target mob past the blocks in
their vicinity (blocks which can be on different chunks).

Requires:
*   R to block data.
*   R/W to individual mob data.
*   R to player data.


Mob and player physics
----------------------

Unlike items, mobs (and players) can collide with each other, block passage,
push off cliffs, etc.

Requires:
*   R to block data.
*   R/W to individual mob data.
*   R/W to other mob data.
*   R to player data.


Player block interactions
-------------------------

This includes digging, placing blocks and other interactions (workbench
crafting, chests, furnace operation, etc.).

Requires:
*   R/W to block data.
*   R/W to player inventory.
*   W to item data. (blocks removed may spawn items)


Mob block interactions
----------------------

Creepers can 'interact' with blocks by exploding. Potentially we might want to
have zombies able to beat down walls to come after players in some kind of
zombie defence mod.

Requires:
*   R/W to block data.
*   W to item data. (blocks removed may spawn items)


Player item pickup
------------------

Requires:
*   R/W to item data.
*   R/W to player inventory.
