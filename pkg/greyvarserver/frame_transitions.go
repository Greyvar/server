package greyvarserver

import (
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	"github.com/greyvar/server/pkg/worlds"
	log "github.com/sirupsen/logrus"
)

func (s *serverInterface) transitionPlayerToGrid(rp *RemotePlayer, destGridId string, x int32, y int32) bool {
	world := s.worldForPlayer(rp)
	if world == nil {
		return false
	}

	if _, ok := world.Grids[destGridId]; !ok {
		return false
	}

	oldGridId := rp.CurrentGridId
	oldWorldId := rp.CurrentWorldId

	rp.PendingGridTransition = nil
	if dx, dy, ok := worlds.ScrollDeltaBetween(world, oldGridId, destGridId); ok && (dx != 0 || dy != 0) {
		rp.PendingGridTransition = &GridTransitionInfo{
			FromGridId:   oldGridId,
			ScrollDeltaX: dx,
			ScrollDeltaY: dy,
		}
	}

	for _, other := range s.remotePlayers {
		if other == rp {
			continue
		}

		if other.CurrentWorldId != oldWorldId || other.CurrentGridId != oldGridId {
			continue
		}

		if _, known := other.KnownEntities[rp.Entity.ServerId]; known {
			other.PendingDespawns = append(other.PendingDespawns, rp.Entity.ServerId)
			delete(other.KnownEntities, rp.Entity.ServerId)
		}
	}

	for id := range rp.KnownEntities {
		rp.PendingDespawns = append(rp.PendingDespawns, id)
	}

	rp.CurrentGridId = destGridId
	rp.Entity.GridId = destGridId
	rp.Entity.WorldId = rp.CurrentWorldId
	rp.Entity.X = x
	rp.Entity.Y = y
	rp.NeedsGridUpdate = true

	log.WithFields(log.Fields{
		"player":   rp.Username,
		"entityId": rp.Entity.ServerId,
		"world":    rp.CurrentWorldId,
		"fromGrid": oldGridId,
		"toGrid":   destGridId,
		"x":        x,
		"y":        y,
	}).Info("Player transitioned grid")

	return true
}

func (s *serverInterface) tryEdgeTransition(rp *RemotePlayer, mr *pb.MoveRequest) bool {
	grid := s.gridForPlayer(rp)
	world := s.worldForPlayer(rp)
	if grid == nil || world == nil {
		return false
	}

	ent := rp.Entity
	row, col := tileCoordsFromPixels(ent.X, ent.Y)

	// Movement deltas: mr.X = col axis, mr.Y = row axis (matches WASD / arrow keys).
	if mr.Y < 0 && row == 0 {
		destId, ok := worlds.AdjacentGridId(world, rp.CurrentGridId, -1, 0)
		if !ok {
			return false
		}

		dest := world.Grids[destId]
		entryX := ent.X
		entryY := int32((dest.RowCount - 1) * tileSizePx)
		return s.transitionPlayerToGrid(rp, destId, entryX, entryY)
	}

	if mr.Y > 0 && row >= grid.RowCount-1 {
		destId, ok := worlds.AdjacentGridId(world, rp.CurrentGridId, 1, 0)
		if !ok {
			return false
		}

		entryX := ent.X
		entryY := int32(0)
		return s.transitionPlayerToGrid(rp, destId, entryX, entryY)
	}

	if mr.X < 0 && col == 0 {
		destId, ok := worlds.AdjacentGridId(world, rp.CurrentGridId, 0, -1)
		if !ok {
			return false
		}

		dest := world.Grids[destId]
		entryX := int32((dest.ColCount - 1) * tileSizePx)
		entryY := ent.Y
		return s.transitionPlayerToGrid(rp, destId, entryX, entryY)
	}

	if mr.X > 0 && col >= grid.ColCount-1 {
		destId, ok := worlds.AdjacentGridId(world, rp.CurrentGridId, 0, 1)
		if !ok {
			return false
		}

		entryX := int32(0)
		entryY := ent.Y
		return s.transitionPlayerToGrid(rp, destId, entryX, entryY)
	}

	return false
}

func (s *serverInterface) tryTeleportTile(rp *RemotePlayer) bool {
	grid := s.gridForPlayer(rp)
	if grid == nil {
		return false
	}

	row, col := tileCoordsFromPixels(rp.Entity.X, rp.Entity.Y)
	if row >= grid.RowCount || col >= grid.ColCount {
		return false
	}

	rowTiles, ok := grid.Tiles[row]
	if !ok {
		return false
	}

	tile := rowTiles[col]
	if tile == nil || tile.TeleportDst == "" {
		return false
	}

	x := int32(tile.TeleportX * tileSizePx)
	y := int32(tile.TeleportY * tileSizePx)

	return s.transitionPlayerToGrid(rp, tile.TeleportDst, x, y)
}
