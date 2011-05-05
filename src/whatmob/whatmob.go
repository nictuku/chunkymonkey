// Utility to analyze an intercept log and look for unknown mobs metadata.
package main

import (
	"encoding/line"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"chunkymonkey/mob"
	"chunkymonkey/types"
)

// TODO: Hook into intercept as an optional feature, so we can get live
// warnings when unknown mob/metadata is seen in a notchian server. Would also
// be able to drop the log parsing via regexp.

var interceptLog = flag.String(
	"interceptLog", "intercept.log",
	"The log file from an intercept session with interesting mobs activity.")

const maxLength = 400

var packetObjectSpawnRegexp = regexp.MustCompile(
	// 2011/05/01 19:37:05.648176 (S->C) PacketEntitySpawn(entityId=9272, mobType=91, position=&{3742 2496 -7321}, look=&{0 0}, metadata=[{0 0 0} {0 16 0}])
	`.*\(S->C\) PacketEntitySpawn\(entityId=[0-9]+, mobType=([0-9]+), position=[^,]+, look=[^,]+, metadata=(.*)\)$`)

func whatMob(logline []byte) os.Error {
	x := packetObjectSpawnRegexp.FindAllSubmatch(logline, 2)
	if x == nil {
		return nil
	}

	t, metadata := string(x[0][1]), string(x[0][2])
	mobType, err := strconv.Atoi(t)
	if err != nil {
		return err
	}
	fmt.Println(mob.Mobs[types.EntityMobType(mobType)].Name, string(metadata))
	return nil
}

func main() {
	f, err := os.Open(*interceptLog)
	if err != nil {
		exit(err)
	}

	l := line.NewReader(f, maxLength)

	for {
		x, tooLong, err := l.ReadLine()
		if err == os.EOF {
			break
		}
		if err != nil {
			exit(err)
		}
		if tooLong {
			// Skip this line. The next result from ReadLine() will
			// be a partial line, but as long as it doesn't match
			// our regular expression we're fine.
			continue
		}
		if err = whatMob(x); err != nil {
			fmt.Println(err)
		}
	}
}

func exit(err os.Error) {
	fmt.Println(err)
	os.Exit(1)
}
