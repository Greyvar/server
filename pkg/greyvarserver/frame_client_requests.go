package greyvarserver;

import (
	pb "github.com/greyvar/server/pkg/greyvarproto"
	log "github.com/sirupsen/logrus"
)


func processMoveRequest(s *serverInterface, mr *pb.MoveRequest) {
	log.WithFields(log.Fields{
		"MR": mr,
	}).Info("MoveRequest");

	s.remotePlayers[int64(mr.PlayerId)].Entity.X += (mr.X * 2);
	s.remotePlayers[int64(mr.PlayerId)].Entity.Y += (mr.Y * 2);
}

func processClientRequests(s *serverInterface, req *pb.ClientRequests) {
	if req.MoveRequest != nil {
		processMoveRequest(s, req.MoveRequest);
	}
}


