package greyvarserver;

import (
	log "github.com/sirupsen/logrus"
	pb "github.com/greyvar/server/gen/greyvarprotocol"
)

func frameNewEntdefs(s *serverInterface, p *RemotePlayer) {
	for name, _ := range s.entityDefinitions {
		if _, ok := p.KnownEntdefs[name]; !ok {
			log.Infof("Need to tell %v about entdef %v", p.Username, name)

			serverEntdef := s.entityDefinitions[name]

			netEntdef := &pb.EntityDefinition{
				Name: serverEntdef.Title,
				Texture: serverEntdef.Texture,
			}

			for name, serverEntState := range serverEntdef.States {
				netEntState := &pb.EntityState {
					Frames: serverEntState.Frames,
					Name: name, // serverEntState.Name is not used really
				}


				netEntdef.States = append(netEntdef.States, netEntState)
			}

			p.currentFrame.EntityDefinitions = append(p.currentFrame.EntityDefinitions, netEntdef)

			p.KnownEntdefs[name] = true;
		}
	}
}
