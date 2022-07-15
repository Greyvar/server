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
	gridToSend := new(pb.Grid);

	for _, memTile := range s.grids[0].Tiles {
		netTile := new(pb.Tile);
		netTile.Row = memTile.Y;
		netTile.Col = memTile.X;
		netTile.Tex = memTile.Texture;
		netTile.Rot = memTile.Rot
		netTile.FlipH = memTile.FlipH
		netTile.FlipV = memTile.FlipV
		gridToSend.Tiles = append(gridToSend.Tiles, netTile);
	}

	return gridToSend;
}
