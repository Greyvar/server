package greyvarserver;

import (
	"context"
	pb "github.com/greyvar/server/pkg/greyvarproto"
)

func (s *serverInterface) GetServerFrame(ctx context.Context, req *pb.ClientRequests) (*pb.ServerFrameResponse, error) {
	res := new(pb.ServerFrameResponse);

	FrameGridUpdates(s, res)

	for i, _ := range s.remotePlayers {
		FramePlayerSpawns(s, res, s.remotePlayers[i])
	}

	return res, nil;
}
