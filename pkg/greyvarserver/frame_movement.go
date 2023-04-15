package greyvarserver;

import (
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
	"math"
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

	checkEntityCollisions(s, rp)
}

func checkEntityCollisions(s *serverInterface, rp *RemotePlayer) {
	for _, ent := range s.entityInstances {
		if ent.Definition == "player" { continue }

		dX := math.Abs(float64(rp.Entity.X - ent.X))
		dY := math.Abs(float64(rp.Entity.Y - ent.Y))

		if dX < 12 && dY < 12 {
			triggerEntityCollection(rp, ent)
		}
	}
}

func triggerEntityCollection(rp *RemotePlayer, ent *Entity) {
	log.Infof("Collission %v %v", rp, ent)

	ent.State = "pressed"

	esc := &pb.EntityStateChange {
		EntityId: ent.ServerId,
		NewState: ent.State,
	}

	rp.currentFrame.EntityStateChanges = append(rp.currentFrame.EntityStateChanges, esc)
}
