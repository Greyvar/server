package greyvarserver;

import (
	"time"
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
	"nhooyr.io/websocket/wspb"
	"context"
)

func (s *serverInterface) frame() {
	s.frameTime = time.Now().UnixNano();

	s.processPlayerRequests();

	s.createServerUpdates();

	s.sendServerUpdates();

	//log.Infof("frame")
}

func (s *serverInterface) processPlayerRequests() {
	for _, p := range s.remotePlayers {
		p.currentFrame = &pb.ServerUpdate{}


		for len(p.pendingRequests) > 0 {
			req := p.pendingRequests[0]
			p.pendingRequests = p.pendingRequests[1:]

			if req.MoveRequest != nil {
				processMoveRequest(s, p, req.MoveRequest);
			}
		}

	}
}

func (s *serverInterface) createServerUpdates() {
	for _, p := range s.remotePlayers {
		frameNewEntdefs(s, p)
		frameSpawnPlayer(s, p)
		frameSpawnEntities(s, p)
		frameGridUpdates(s, p)
		frameEntityPositions(s, p);
	}
}

func (s *serverInterface) sendServerUpdates() {
	// We deliberatly iterate over remotePlayers again here, just sending
	// the server frame - so that hopefully clients don't perceive lag from 
	// processing the frame updates for each player, above.
	for _, player := range s.remotePlayers {
		s.sendServerFrameForPlayer(player)
	}
}

func (s *serverInterface) sendServerFrameForPlayer(p *RemotePlayer) {
	s.sendServerFrame(p.currentFrame, p)
}

func (s *serverInterface) sendServerFrame(frame *pb.ServerUpdate, p *RemotePlayer) {
	err := wspb.Write(context.Background(), p.Connection, frame)

	if err != nil {
		log.Errorf("Could not marshal obj to protobuf in sendMessage: %v", err);
		return
	}
}


func frameEntityPositions(s *serverInterface, p *RemotePlayer) {
	for _, ent := range s.entityInstances {
		entpos := pb.EntityPosition {
			X: ent.X,
			Y: ent.Y,
			EntityId: ent.ServerId,
		}

		p.currentFrame.EntityPositions = append(p.currentFrame.EntityPositions, &entpos);
	}
}

