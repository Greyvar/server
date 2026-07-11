package greyvarserver;

import (
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
	"math"
)

const DISTANCE_PER_REQUEST = 4

func processMoveRequest(s *serverInterface, rp *RemotePlayer, mr *pb.MoveRequest) {
	fields := log.Fields{
		"player":   rp.Username,
		"entityId": rp.Entity.ServerId,
		"gridId":   rp.CurrentGridId,
		"fromX":    rp.Entity.X,
		"fromY":    rp.Entity.Y,
		"deltaX":   mr.X,
		"deltaY":   mr.Y,
	}

	if ((s.frameTime - rp.TimeOfLastMoveRequest) < 100) {
		log.WithFields(fields).Info("MoveRequest throttled")
		return
	}

	if mr.X < -1 || mr.X > 1 || mr.Y < -1 || mr.Y > 1 {
		log.WithFields(fields).Info("MoveRequest rejected: delta out of range")
		return
	}

	if mr.X != 0 && mr.Y != 0 {
		log.WithFields(fields).Info("MoveRequest rejected: diagonal move")
		return
	}

	newX := rp.Entity.X + (mr.X * DISTANCE_PER_REQUEST)
	newY := rp.Entity.Y + (mr.Y * DISTANCE_PER_REQUEST)
	fields["toX"] = newX
	fields["toY"] = newY

	if reason := s.moveBlockReason(rp, rp.Entity, newX, newY); reason != "" {
		if s.tryEdgeTransition(rp, mr) {
			rp.TimeOfLastMoveRequest = s.frameTime
			log.WithFields(fields).Info("MoveRequest triggered grid transition")
			return
		}

		log.WithFields(fields).WithField("reason", reason).Info("MoveRequest blocked")
		rp.TimeOfLastMoveRequest = s.frameTime
		return
	}

	rp.Entity.X = newX
	rp.Entity.Y = newY
	rp.TimeOfLastMoveRequest = s.frameTime

	log.WithFields(fields).Info("MoveRequest accepted")

	checkEntityCollisions(s, rp)

	if s.tryTeleportTile(rp) {
		log.WithFields(fields).Info("MoveRequest triggered teleport tile")
	}
}

func checkEntityCollisions(s *serverInterface, rp *RemotePlayer) {
	for _, ent := range s.entitiesOnGrid(rp.CurrentWorldId, rp.CurrentGridId) {
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
