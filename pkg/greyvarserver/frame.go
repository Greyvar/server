package greyvarserver;

import (
	"context"
	pb "github.com/greyvar/server/pkg/greyvarproto"
	log "github.com/sirupsen/logrus"
)

func processMoveRequest(mr *pb.MoveRequest) {
	log.Info(mr)
}

func (s *serverInterface) GetServerFrame(ctx context.Context, req *pb.ClientRequests) (*pb.ServerFrameResponse, error) {
	res := new(pb.ServerFrameResponse);

	FrameGridUpdates(s, res)

	for i, _ := range s.remotePlayers {
		FramePlayerSpawns(s, res, &s.remotePlayers[i])
	}

	if req.MoveRequest != nil {
		processMoveRequest(req.MoveRequest);
	}

	return res, nil;
}
