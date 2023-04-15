package greyvarserver;

import (
	pb "github.com/greyvar/server/gen/greyvarprotocol"
)

func frameGridUpdates(s *serverInterface, p *RemotePlayer) {
	if (p.NeedsGridUpdate) {
		p.currentFrame.Grid = generateGridUpdate(s);
		p.NeedsGridUpdate = false
	}
}

func generateGridUpdate(s *serverInterface) (*pb.Grid) {
	memGrid := s.grids[0]

	gridToSend := &pb.Grid {
		RowCount: memGrid.RowCount,
		ColCount: memGrid.ColCount,
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
		gridToSend.Tiles = append(gridToSend.Tiles, netTile);
	}

	return gridToSend;
}
