package greyvarserver;

import (
	pb "github.com/greyvar/server/pkg/greyvarproto"
)

func FrameGridUpdates(s *serverInterface, serverFrame *pb.ServerFrameResponse) {
	if isGridUpdateNeeded(s) {
		serverFrame.Grid = generateGridUpdate(s);
	}
}

func isGridUpdateNeeded(s *serverInterface) bool {
	ret := false

	for i, rp := range s.remotePlayers {
		if (rp.NeedsGridUpdate) {
			s.remotePlayers[i].NeedsGridUpdate = false

			ret = true
		}
	}

	return ret
}

func generateGridUpdate(s *serverInterface) (*pb.Grid) {
	gridToSend := new(pb.Grid);

	for _, memTile := range s.grids[0].Tiles {
		netTile := new(pb.Tile);
		netTile.Row = memTile.X;
		netTile.Col = memTile.Y;
		netTile.Tex = memTile.Texture;
		netTile.Rot = memTile.Rot
		netTile.FlipH = memTile.FlipH
		netTile.FlipV = memTile.FlipV
		gridToSend.Tiles = append(gridToSend.Tiles, netTile);
	}

	return gridToSend;
}
