package worlds

import (
	"strconv"
	"strings"

	"github.com/greyvar/datlib/gridfiles"
	"github.com/greyvar/datlib/worldfiles"
)

type World = worldfiles.World
type Summary = worldfiles.Summary
type MapPlacement = worldfiles.MapPlacement

func LoadWorld(name string) (*World, error) {
	return worldfiles.LoadWorld(name)
}

func ListWorlds() ([]Summary, error) {
	return worldfiles.ListWorlds()
}

func ParseGridCoords(gridId string) (row int, col int, ok bool) {
	stem := strings.TrimSuffix(gridId, ".grid")
	stem = strings.TrimSuffix(stem, ".tmj")
	parts := strings.Split(stem, ".")
	if len(parts) < 2 {
		return 0, 0, false
	}

	row, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}

	col, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, false
	}

	return row, col, true
}

func AdjacentGridId(world *World, gridId string, deltaRow int, deltaCol int) (string, bool) {
	row, col, ok := ParseGridCoords(gridId)
	if !ok {
		return "", false
	}

	targetStem := strconv.Itoa(row+deltaRow) + "." + strconv.Itoa(col+deltaCol)

	ext := ".grid"
	if strings.HasSuffix(gridId, ".tmj") {
		ext = ".tmj"
	}

	target := targetStem + ext
	if _, exists := world.Grids[target]; exists {
		return target, true
	}

	altExt := ".tmj"
	if ext == ".tmj" {
		altExt = ".grid"
	}
	alt := targetStem + altExt
	if _, exists := world.Grids[alt]; exists {
		return alt, true
	}

	return "", false
}

// ScrollDeltaBetween returns the pixel offset from fromGridId to toGridId using
// world map placements when available, otherwise row.col grid dimensions.
func ScrollDeltaBetween(world *World, fromGridId string, toGridId string) (int32, int32, bool) {
	if world == nil {
		return 0, 0, false
	}

	from, okFrom := PlacementForGrid(world, fromGridId)
	to, okTo := PlacementForGrid(world, toGridId)
	if !okFrom || !okTo {
		return 0, 0, false
	}

	return int32(to.X - from.X), int32(to.Y - from.Y), true
}

// PlacementForGrid resolves a grid's pixel placement within the world.
func PlacementForGrid(world *World, gridId string) (MapPlacement, bool) {
	return worldfiles.PlacementForGrid(world.Definition, world.Grids, gridId)
}

// GridAt returns a loaded grid by filename (.grid or .tmj).
func GridAt(world *World, gridId string) (*gridfiles.Grid, bool) {
	grid, ok := world.Grids[gridId]
	return grid, ok
}
