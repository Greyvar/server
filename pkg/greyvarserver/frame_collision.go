package greyvarserver

import (
	"github.com/greyvar/datlib/gridfiles"
)

const tileSizePx = 16
const entitySizePx = 16

func isTileTraversable(tile *gridfiles.Tile) bool {
	if tile == nil {
		return false
	}

	if isBlockingTileTexture(tile.Texture) {
		return false
	}

	if tile.Traversable {
		return true
	}

	return true
}

func tileCoordsFromPixels(x int32, y int32) (uint32, uint32) {
	// Grid yaml row/col match screen space: x = col, y = row (see client GridScene).
	row := uint32(y / tileSizePx)
	col := uint32(x / tileSizePx)

	return row, col
}

func (s *serverInterface) gridForPlayer(rp *RemotePlayer) *gridfiles.Grid {
	if rp == nil {
		return nil
	}

	return s.gridById(rp.CurrentWorldId, rp.CurrentGridId)
}

func (s *serverInterface) tileAtGrid(grid *gridfiles.Grid, x int32, y int32) *gridfiles.Tile {
	if grid == nil {
		return nil
	}

	row, col := tileCoordsFromPixels(x, y)

	if row >= grid.RowCount || col >= grid.ColCount {
		return nil
	}

	rowTiles, ok := grid.Tiles[row]
	if !ok {
		return nil
	}

	return rowTiles[col]
}

func (s *serverInterface) tileAtForPlayer(rp *RemotePlayer, x int32, y int32) *gridfiles.Tile {
	return s.tileAtGrid(s.gridForPlayer(rp), x, y)
}

func (s *serverInterface) tileAtGridCell(grid *gridfiles.Grid, row uint32, col uint32) *gridfiles.Tile {
	if grid == nil {
		return nil
	}

	if row >= grid.RowCount || col >= grid.ColCount {
		return nil
	}

	rowTiles, ok := grid.Tiles[row]
	if !ok {
		return nil
	}

	return rowTiles[col]
}

func (s *serverInterface) isTraversableAt(rp *RemotePlayer, x int32, y int32) bool {
	grid := s.gridForPlayer(rp)
	if grid == nil {
		return false
	}

	minRow := uint32(y / tileSizePx)
	maxRow := uint32((y + entitySizePx - 1) / tileSizePx)
	minCol := uint32(x / tileSizePx)
	maxCol := uint32((x + entitySizePx - 1) / tileSizePx)

	for row := minRow; row <= maxRow; row++ {
		for col := minCol; col <= maxCol; col++ {
			if !isTileTraversable(s.tileAtGridCell(grid, row, col)) {
				return false
			}
		}
	}

	return true
}

func (s *serverInterface) isTraversableForPlayer(rp *RemotePlayer, x int32, y int32) bool {
	return s.isTraversableAt(rp, x, y)
}

func (s *serverInterface) isBlockedByPlayer(rp *RemotePlayer, ent *Entity, newX int32, newY int32) bool {
	for _, other := range s.entityInstances {
		if other.ServerId == ent.ServerId || other.Definition != "player" {
			continue
		}

		if other.GridId != rp.CurrentGridId || other.WorldId != rp.CurrentWorldId {
			continue
		}

		if other.X == newX && other.Y == newY {
			return true
		}
	}

	return false
}

func (s *serverInterface) moveBlockReason(rp *RemotePlayer, ent *Entity, newX int32, newY int32) string {
	if !s.isTraversableAt(rp, newX, newY) {
		grid := s.gridForPlayer(rp)
		texture := ""
		if grid != nil {
			row, col := tileCoordsFromPixels(newX, newY)
			if tile := s.tileAtGridCell(grid, row, col); tile != nil {
				texture = tile.Texture
			}
		}
		return "tile blocked (" + texture + ")"
	}

	if s.isBlockedByPlayer(rp, ent, newX, newY) {
		return "player at destination"
	}

	return ""
}

func (s *serverInterface) canMoveTo(rp *RemotePlayer, ent *Entity, newX int32, newY int32) bool {
	return s.moveBlockReason(rp, ent, newX, newY) == ""
}

func (s *serverInterface) entitiesOnGrid(worldId string, gridId string) []*Entity {
	out := make([]*Entity, 0)

	for _, ent := range s.entityInstances {
		if ent.GridId == gridId && ent.WorldId == worldId {
			out = append(out, ent)
		}
	}

	return out
}
