package greyvarserver;

import (
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

const DISTANCE_PER_REQUEST = 4

func processMoveRequest(s *serverInterface, rp *RemotePlayer, mr *pb.MoveRequest) {
	if ((s.frameTime - rp.TimeOfLastMoveRequest) < 100) {
		return // Movement key spam
	}

	log.WithFields(log.Fields{
		"MR": mr,
		"RP": rp.Username,
		"ServerEntId": rp.Entity.ServerId,
	}).Info("MoveRequest");

	// X or Y should be 1 or -1
	// Not absolute grid units

	rp.Entity.X += (mr.X * DISTANCE_PER_REQUEST);
	rp.Entity.Y += (mr.Y * DISTANCE_PER_REQUEST);
	rp.TimeOfLastMoveRequest = s.frameTime;
}

// FIXME ClientRequests should be queued and handled in frame() !!
func (server *serverInterface) handleClientRequests(rp *RemotePlayer, req *pb.ClientRequests) {
	if req.MoveRequest != nil {
		processMoveRequest(server, rp, req.MoveRequest);
	}
}


