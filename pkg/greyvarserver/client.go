package greyvarserver;

import (
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)


func processMoveRequest(s *serverInterface, rp *RemotePlayer, mr *pb.MoveRequest) {
	log.WithFields(log.Fields{
		"MR": mr,
		"RP": rp.Username,
		"EntId": rp.Entity.Id,
	}).Info("MoveRequest");

	rp.Entity.X += (mr.X * 2);
	rp.Entity.Y += (mr.Y * 2);
}

func (server *serverInterface) handleClientRequests(rp *RemotePlayer, req *pb.ClientRequests) {
	if req.MoveRequest != nil {
		processMoveRequest(server, rp, req.MoveRequest);
	}
}


