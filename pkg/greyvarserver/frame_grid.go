package greyvarserver;

import (
	pb "github.com/greyvar/server/gen/greyvarprotocol"
)

func frameGridUpdates(s *serverInterface, p *RemotePlayer) {
	if p.NeedsGridUpdate {
		p.currentFrame.Grid = generateGridUpdate(s, p)
		p.KnownEntities = make(map[int64]*Entity)
		p.KnownEntdefs = make(map[string]bool)
		p.NeedsGridUpdate = false
	}
}

func generateGridUpdate(s *serverInterface, p *RemotePlayer) (*pb.Grid) {
	worldId := p.CurrentWorldId
	gridId := p.CurrentGridId
	memGrid := s.gridById(worldId, gridId)
	if memGrid == nil {
		return nil
	}

	gridToSend := &pb.Grid {
		Title: memGrid.Filename,
		GridId: gridId,
		WorldId: worldId,
		RowCount: memGrid.RowCount,
		ColCount: memGrid.ColCount,
	}

	for _, tileset := range memGrid.Tilesets {
		gridToSend.Tilesets = append(gridToSend.Tilesets, &pb.Tileset{
			Key:        tileset.Key,
			Image:      tileset.ImagePath,
			TileWidth:  uint32(tileset.TileWidth),
			TileHeight: uint32(tileset.TileHeight),
			Columns:    uint32(tileset.Columns),
		})
	}

	if p.PendingGridTransition != nil {
		gridToSend.Transition = &pb.GridTransition{
			FromGridId:   p.PendingGridTransition.FromGridId,
			ScrollDeltaX: p.PendingGridTransition.ScrollDeltaX,
			ScrollDeltaY: p.PendingGridTransition.ScrollDeltaY,
		}
		p.PendingGridTransition = nil
	}
	
	for _, pos := range memGrid.CellIterator() {
		memTile := memGrid.Tiles[pos.Row][pos.Col]

		netTile := new(pb.Tile);
		netTile.Row = memTile.Row;
		netTile.Col = memTile.Col;
		netTile.Tex = memTile.Texture;
		netTile.Rot = memTile.Rot
		netTile.FlipH = memTile.FlipH
		netTile.FlipV = memTile.FlipV
		netTile.AtlasKey = memTile.AtlasKey
		netTile.FrameIndex = memTile.FrameIndex
		gridToSend.Tiles = append(gridToSend.Tiles, netTile);
	}

	return gridToSend;
}
