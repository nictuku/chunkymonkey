package gamerules

import (
	"reflect"
	"strings"
	"testing"
)

const threeRecipes = ("[\n" +
	"  {\n" +
	"    \"Comment\": \"log->planks\",\n" +
	"    \"Input\": [\n" +
	"      \"L\"\n" +
	"    ],\n" +
	"    \"InputTypes\": {\n" +
	"      \"L\": [\n" +
	"        {\"Id\": 17, \"Data\": 0},\n" +
	"        {\"Id\": 17, \"Data\": 1},\n" +
	"        {\"Id\": 17, \"Data\": 2}\n" +
	"      ]\n" +
	"    },\n" +
	"    \"OutputTypes\": [\n" +
	"      {\"Id\": 5},\n" +
	"      {\"Id\": 5},\n" +
	"      {\"Id\": 5}\n" +
	"    ],\n" +
	"    \"OutputCount\": 4\n" +
	"  },\n" +
	"  {\n" +
	"    \"Comment\": \"TNT\",\n" +
	"    \"Input\": [\n" +
	"      \"GSG\",\n" +
	"      \"SGS\",\n" +
	"      \"GSG\"\n" +
	"    ],\n" +
	"    \"InputTypes\": {\n" +
	"      \"G\": [{\"Id\": 289}],\n" +
	"      \"S\": [{\"Id\": 12}]\n" +
	"    },\n" +
	"    \"OutputTypes\": [{\"Id\": 46}],\n" +
	"    \"OutputCount\": 1\n" +
	"  },\n" +
	"  {\n" +
	"    \"Comment\": \"flint and steel\",\n" +
	"    \"Input\": [\n" +
	"      \"F \",\n" +
	"      \" I\"\n" +
	"    ],\n" +
	"    \"InputTypes\": {\n" +
	"      \"I\": [{\"Id\": 265}],\n" +
	"      \"F\": [{\"Id\": 318}]\n" +
	"    },\n" +
	"    \"OutputTypes\": [{\"Id\": 259}],\n" +
	"    \"OutputCount\": 1\n" +
	"  }\n" +
	"]\n")

func createItemTypes() (items ItemTypeMap) {
	// The information in the ItemType isn't actually used, so we only need
	// them to exist. We do set the ID so that the tests can easily check for
	// equality.
	items = ItemTypeMap{
		5:   &ItemType{},
		12:  &ItemType{},
		17:  &ItemType{},
		46:  &ItemType{},
		259: &ItemType{},
		265: &ItemType{},
		289: &ItemType{},
		318: &ItemType{},
	}
	for id := range items {
		items[id].Id = id
	}

	return
}

func assertRecipesEq(t *testing.T, expected, result *Recipe) {
	if !reflect.DeepEqual(expected, result) {
		t.Error("Recipes differed.")
		t.Errorf("Expected: %#v", expected)
		t.Errorf("Result:   %#v", result)
	}
}

func TestLoadRecipes(t *testing.T) {
	itemTypes := createItemTypes()
	t.Logf(threeRecipes)
	reader := strings.NewReader(threeRecipes)

	recipes, err := LoadRecipes(reader, itemTypes)

	if err != nil {
		t.Fatalf("Expected no error loading recipes, got: %v", err)
	}
	if recipes == nil {
		t.Fatal("Got nil *RecipeSet")
	}
	if len(recipes.recipes) != 5 {
		t.Fatalf("Expected 5 recipes, got: %d", len(recipes.recipes))
	}

	// We expect to see:
	// normal logs to planks
	assertRecipesEq(
		t,
		&Recipe{
			Comment: "log->planks",
			Width:   1,
			Height:  1,
			Input: []Slot{
				{17, 0, 0},
			},
			Output: Slot{5, 4, 0},
		},
		&recipes.recipes[0],
	)
	// spruce logs to planks
	assertRecipesEq(
		t,
		&Recipe{
			Comment: "log->planks",
			Width:   1,
			Height:  1,
			Input: []Slot{
				{17, 0, 1},
			},
			Output: Slot{5, 4, 0},
		},
		&recipes.recipes[1],
	)
	// birch logs to planks
	assertRecipesEq(
		t,
		&Recipe{
			Comment: "log->planks",
			Width:   1,
			Height:  1,
			Input: []Slot{
				{17, 0, 2},
			},
			Output: Slot{5, 4, 0},
		},
		&recipes.recipes[2],
	)
	// TNT
	assertRecipesEq(
		t,
		&Recipe{
			Comment: "TNT",
			Width:   3,
			Height:  3,
			Input: []Slot{
				{289, 0, 0},
				{12, 0, 0},
				{289, 0, 0},
				{12, 0, 0},
				{289, 0, 0},
				{12, 0, 0},
				{289, 0, 0},
				{12, 0, 0},
				{289, 0, 0},
			},
			Output: Slot{46, 1, 0},
		},
		&recipes.recipes[3],
	)
	// flight and steel
	assertRecipesEq(
		t,
		&Recipe{
			Comment: "flint and steel",
			Width:   2,
			Height:  2,
			Input: []Slot{
				{318, 0, 0},
				{0, 0, 0},
				{0, 0, 0},
				{265, 0, 0},
			},
			Output: Slot{259, 1, 0},
		},
		&recipes.recipes[4],
	)
}

func assertLoadError(t *testing.T, input string) {
	itemTypes := createItemTypes()
	reader := strings.NewReader(input)

	_, err := LoadRecipes(reader, itemTypes)

	if err == nil {
		t.Errorf("Should have got error loading: %s", input)
	} else {
		t.Logf("Correctly got error: %v", err)
	}
}

func TestLoadRecipes_Errors(t *testing.T) {
	assertLoadError(t, ",")
	// Too high.
	assertLoadError(t, "[\n"+
		"  {\n"+
		"    \"Comment\": \"log->planks\",\n"+
		"    \"Input\": [\n"+
		"      \"L\", \"L\", \"L\", \"L\"\n"+
		"    ],\n"+
		"    \"InputTypes\": {\n"+
		"      \"L\": [{\"Id\": 17}]\n"+
		"    },\n"+
		"    \"OutputTypes\": [{\"Id\": 5}],\n"+
		"    \"OutputCount\": 4\n"+
		"  }\n"+
		"]")
	// Too wide.
	assertLoadError(t, "[\n"+
		"  {\n"+
		"    \"Comment\": \"log->planks\",\n"+
		"    \"Input\": [\n"+
		"      \"LLLL\"\n"+
		"    ],\n"+
		"    \"InputTypes\": {\n"+
		"      \"L\": [{\"Id\": 17}]\n"+
		"    },\n"+
		"    \"OutputTypes\": [{\"Id\": 5}],\n"+
		"    \"OutputCount\": 4\n"+
		"  }\n"+
		"]")
	// Irregular input rows.
	assertLoadError(t, "[\n"+
		"  {\n"+
		"    \"Comment\": \"log->planks\",\n"+
		"    \"Input\": [\n"+
		"      \"LLL\",\n"+
		"      \"LL\"\n"+
		"    ],\n"+
		"    \"InputTypes\": {\n"+
		"      \"L\": [{\"Id\": 17}]\n"+
		"    },\n"+
		"    \"OutputTypes\": [{\"Id\": 5}],\n"+
		"    \"OutputCount\": 4\n"+
		"  }\n"+
		"]")
	// Differing counts of input/output types.
	assertLoadError(t, "[\n"+
		"  {\n"+
		"    \"Comment\": \"log->planks\",\n"+
		"    \"Input\": [\n"+
		"      \"L\"\n"+
		"    ],\n"+
		"    \"InputTypes\": {\n"+
		"      \"L\": [{\"Id\": 17},{\"Id\": 18}]\n"+
		"    },\n"+
		"    \"OutputTypes\": [{\"Id\": 5}],\n"+
		"    \"OutputCount\": 4\n"+
		"  }\n"+
		"]")
	// Bad key name.
	assertLoadError(t, "[\n"+
		"  {\n"+
		"    \"Comment\": \"log->planks\",\n"+
		"    \"Input\": [\n"+
		"      \"L\"\n"+
		"    ],\n"+
		"    \"InputTypes\": {\n"+
		"      \"L\": [{\"Id\": 17}],\n"+
		"      \"LONGNAME\": [{\"Id\": 18}]\n"+
		"    },\n"+
		"    \"OutputTypes\": [{\"Id\": 5}],\n"+
		"    \"OutputCount\": 4\n"+
		"  }\n"+
		"]")
	// Undefined key name.
	assertLoadError(t, "[\n"+
		"  {\n"+
		"    \"Comment\": \"log->planks\",\n"+
		"    \"Input\": [\n"+
		"      \"L\"\n"+
		"    ],\n"+
		"    \"InputTypes\": {\n"+
		"      \"X\": [{\"Id\": 17}]\n"+
		"    },\n"+
		"    \"OutputTypes\": [{\"Id\": 5}],\n"+
		"    \"OutputCount\": 4\n"+
		"  }\n"+
		"]")
}
