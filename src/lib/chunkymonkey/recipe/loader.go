package recipe

import (
    "fmt"
    "io"
    "json"
    "os"

    "chunkymonkey/itemtype"
    . "chunkymonkey/types"
)

type typeInstance struct {
    Id   ItemTypeId
    Data ItemData
}

func (ti *typeInstance) createRecipeSlot(itemTypes itemtype.ItemTypeMap) (slot RecipeSlot, err os.Error) {
    var ok bool
    slot.Type, ok = itemTypes[ti.Id]
    if !ok {
        err = fmt.Errorf("Item type %d does not exist")
        return
    }
    slot.Data = ti.Data
    return
}

// recipeTemplate is the serialization structure for 0:M Recipes.
type recipeTemplate struct {
    Comment     string
    Input       []string
    InputTypes  map[string][]typeInstance
    OutputTypes []typeInstance
    OutputCount ItemCount
    height      byte
    width       byte
}

// init checks and initialises a recipe template.
func (rt *recipeTemplate) init() (err os.Error) {
    // Check width/height.
    height := len(rt.Input)
    if height < 1 || height > maxRecipeHeight {
        err = fmt.Errorf("Invalid recipe height (%d) in %q", height, rt.Comment)
        return
    }
    width := len(rt.Input[0])
    if width < 1 || width > maxRecipeWidth {
        err = fmt.Errorf("Invalid recipe width (%d) in %q", width, rt.Comment)
        return
    }
    rt.height = byte(width)
    rt.width = byte(height)

    // Check for irregular input row sizes.
    for _, row := range rt.Input {
        if len(row) != width {
            err = fmt.Errorf("Irregular recipe input rows in %q", rt.Comment)
            return
        }
    }

    // Check for differing counts of InputType(s) and OutputType.
    recipeCount := len(rt.OutputTypes)
    for i := range rt.InputTypes {
        if len(rt.InputTypes[i]) != recipeCount {
            err = fmt.Errorf("Irregular input type count in %q", rt.Comment)
            return
        }
        // Check for InputType keys with len() != 1.
        if len(i) != 1 {
            err = fmt.Errorf("Bad input type key %q in %q", i, rt.Comment)
            return
        }
    }

    return
}

// numRecipes returns the number of recipes generated by this template.
func (rt *recipeTemplate) numRecipes() int {
    return len(rt.OutputTypes)
}

// createRecipe creates one of the recipes from the template.
func (rt *recipeTemplate) createRecipe(recipeIndex int, itemTypes itemtype.ItemTypeMap) (recipe Recipe, err os.Error) {

    recipe = Recipe{
        Comment:     rt.Comment,
        Width:       byte(rt.width),
        Height:      byte(rt.height),
        Input:       make([]RecipeSlot, rt.width*rt.height),
        OutputCount: rt.OutputCount,
    }

    slotIndex := 0
    for _, inRow := range rt.Input {
        for _, inSlot := range inRow {
            if inSlot == ' ' {
                recipe.Input[slotIndex] = RecipeSlot{nil, 0}
            } else {
                typeKey := string(inSlot)
                inputTypeSeq, ok := rt.InputTypes[typeKey]
                if !ok {
                    err = fmt.Errorf(
                        "Recipe template %q: Item code %q found in Input which"+
                            " does not exist in InputTypes",
                        rt.Comment, typeKey)
                    return
                }
                recipe.Input[slotIndex], _ = inputTypeSeq[recipeIndex].createRecipeSlot(itemTypes)
                if err != nil {
                    return
                }
            }
            slotIndex++
        }
    }

    recipe.Output, err = rt.OutputTypes[recipeIndex].createRecipeSlot(itemTypes)
    if err != nil {
        return
    }

    return
}

// LoadRecipes reads recipes from a JSON template in reader. itemTypes must be
// provided to map item type IDs to known items.
func LoadRecipes(reader io.Reader, itemTypes itemtype.ItemTypeMap) (recipes *RecipeSet, err os.Error) {
    var templates []recipeTemplate

    decoder := json.NewDecoder(reader)
    err = decoder.Decode(&templates)
    if err != nil {
        return
    }

    // Count up how many recipes there will be generated (to allocate the exact
    // amount of memory required and avoid unnecessary reallocation).
    numRecipes := 0
    for i := range templates {
        tmpl := &templates[i]
        err = tmpl.init()
        if err != nil {
            return
        }
        numRecipes += tmpl.numRecipes()
    }

    recipes = &RecipeSet{
        Recipes: make([]Recipe, numRecipes),
    }

    curRecipe := 0
    for i := range templates {
        tmpl := &templates[i]

        numRecipes := tmpl.numRecipes()
        for recipeIndex := 0; recipeIndex < numRecipes; recipeIndex++ {
            recipes.Recipes[curRecipe], err = tmpl.createRecipe(recipeIndex, itemTypes)
            if err != nil {
                return
            }
            curRecipe++
        }
    }

    return
}

func LoadRecipesFromFile(filename string, itemTypes itemtype.ItemTypeMap) (recipes *RecipeSet, err os.Error) {
    file, err := os.Open(filename)
    if err != nil {
        return
    }
    defer file.Close()

    return LoadRecipes(file, itemTypes)
}
