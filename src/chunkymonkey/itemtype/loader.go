package itemtype

import (
	"io"
	"json"
	"os"
	"strconv"

	. "chunkymonkey/types"
)

func LoadItemDefs(reader io.Reader) (items ItemTypeMap, err os.Error) {
	itemsStr := make(map[string]*ItemType)
	decoder := json.NewDecoder(reader)
	err = decoder.Decode(&itemsStr)

	// Convert map string keys to ints.
	items = make(ItemTypeMap)
	for idStr, item := range itemsStr {
		var id int
		id, err = strconv.Atoi(idStr)

		if err != nil {
			return
		}

		item.Id = ItemTypeId(id)

		items[ItemTypeId(id)] = itemsStr[idStr]
	}

	// Include a "null" item type.
	items[ItemTypeIdNull] = &ItemType{
		Id:       ItemTypeIdNull,
		Name:     "null item",
		MaxStack: 0,
	}

	return
}

func LoadItemTypesFromFile(filename string) (items ItemTypeMap, err os.Error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	return LoadItemDefs(file)
}
