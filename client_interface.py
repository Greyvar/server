import socketserver
import player
import tile
import threading
import game
import socket
import logging, logging.config
import yaml
import math
import time

ETB = "---\n"

class client_interface(socketserver.StreamRequestHandler):
    localPlayers = {}
    templateCommands = {}
    grid = None

    def setup(self):
        self.alive = True
        self.server.game.registerClient(self);

    def handle(self):
        logging.debug("New connection.");

        welc = {
            "serverVersion": "greyvar.devel",
            "name": "^8The Construct"
        }

        self.request.send(ETB.encode("utf-8"))
        self.send("WELC", welc)

        chunkBuf = ""

        while self.alive:
            try:
                chunk = self.request.recv(1024); 
            except socket.error as e:
                logging.debug("socket exception: " + str(e))
                break;

            if not chunk: break;

            chunkBuf += chunk.decode("utf-8")

            print(chunkBuf)

            while ETB in chunkBuf:
                packet, chunkBuf = chunkBuf.split(ETB, 1)

                self.parse_chunk(packet.replace("\0", ""))

        self.server.game.unregisterClient(self)

    def parse_chunk(self, chunk):
        if len(chunk) == 0:
            return

        logging.debug("parse chunk (len: " + str(len(chunk)) + "):\nppp>>>\n" + str(chunk) + "\nppp>>>\n");

        req = yaml.load(chunk)
        cmd = req["command"]

        if cmd != "HELO" and "username" in req: 
          req["player"] = self.server.game.getPlayerByUsername(req["username"])

        if cmd == "HELO":
            self.handle_helo(req)
        elif cmd == "INIT":
            self.handle_init(req);
        elif cmd == "QUIT":
            return;
        elif cmd == "HALT":
            self.handle_halt(req);
        elif cmd == "MOVR":
            self.handle_movr(req);
        else:
            logging.debug("Unknown command from client: " + str(req));

    def subdict(self, d, *args):
        if isinstance(d, object):
            d = d.__dict__

        return dict([i for i in list(d.items()) if i[0] in args])

    def finish(self):
        self.alive = False

    def send_player_you(self, player):
        plru = self.subdict(player, "id", "username", "skin");

        self.send("PLRU", plru)

    def send_player_already_here(self, player):
        plrh = self.subdict(player, "id", "username", "color")

        self.send("PLRH", plrh)

    def send_player_join(self, player):
        logging.debug("sending new player")

        plrj = self.subdict(player, "id", "username", "skin", "color")

        self.send("PLRJ", plrj);
        self.send_move(player)

    def send_player_quit(self, player):
        if player != None:
            self.send("PLRQ", self.subdict(player, "id"))

    def send_spawn(self, player):
        spwn = self.subdict(player, "id", "posX", "posY")

        self.send("SPWN", spwn);

    def send_move(self, player):
        move = {
            "posX": player.posX,
            "posY": player.posY,
            "playerId": player.id,
            "walkState": int(math.floor(player.walkState))
        }

        print(("move", player.posX, player.posY))

        self.send("MOVE", move)

    def register_template_command(self, name, command):
        self.templateCommands[name] = command;

        self.send("TMPL", command)

    def templateize(self, templateName, original):
        template = self.templateCommands[templateName];

        print(("checking ", templateName, " vs ", original))
        
        canOptimize = True

        for key in template:
            if key not in original or original[key] != template[key]:
                canOptimize = False
                break;

        if canOptimize:
            print("can optimize!")
            for key in template:
                original.pop(key, None)

            original["templateRef"] = templateName

        return original

    def send_grid(self):
        self.register_template_command("tileDefault", {
          "rot": 0,
          "vrt": True,
          "hor": True,
        });

        for row, col, tile in self.grid.allTiles():
            tile = {
                "row": row,
                "col": col,
                "tex": tile.tex,
                "rot": tile.rot,
                "vrt": tile.flipV,
                "hor": tile.flipH
            }

            tile = self.templateize("tileDefault", tile)

            self.send("TILE", tile);

        print(self.grid.entities)

        for row, col, ent in self.grid.allEntities():
            if ent == None: continue

            ent = {
              "x": row,
              "y": col,
              "id": ent.id,
              "tex": ent.tex
            }

            self.send("ENT", ent)

    def send(self, command, payload):
        payload["apiVersion"] = "v1"
        payload["command"] = command.strip().upper()
        payload = yaml.dump(payload, default_flow_style = False)
        
        message = payload + ETB;
        self.request.send(message.encode("utf-8"))

    def handle_init(self, init):
        self.send_grid()

        # Tell us who is already here
        for client in list(self.server.game.clientsToPlayers.values()):
            for player in client["players"]:
                self.send_player_already_here(player)
                self.send_spawn(player)

    def handle_helo(self, helo):
        newPlayer = self.server.game.registerPlayer(self, helo["username"])

        if newPlayer == None:
            return

        # Register
        self.localPlayers[helo["username"]] = newPlayer

        # Tell others that we have joined
        for otherClient in list(self.server.game.clientsToPlayers.keys()):
          otherClient.send_player_join(newPlayer)
          otherClient.send_spawn(newPlayer)

        # Tell us who we are
        self.send_player_you(newPlayer)

    def handle_movr(self, movr):
        player = movr["player"]

        if (time.time() - player.timeOfLastMove) < .1:
            self.send("BLKD", {})
            return
        else:
            player.timeOfLastMove = time.time()

        moveX = movr['x']
        moveY = movr['y']

        currentTile = player.getCurrentTile()
        needsTeleport = False

        print(("handleMovr() ", moveX, moveY, currentTile.dstDir, currentTile.dstGrid, player.grid)) 

        if currentTile.dstGrid != None:
            logging.debug("this tile has a destination: " + currentTile.dstDir)
            if moveX > 0 and currentTile.dstDir == "EAST": needsTeleport = True
            if moveX < 0 and currentTile.dstDir == "WEST": needsTeleport = True
            if moveY > 0 and currentTile.dstDir == "NORTH": needsTeleport = True
            if moveY < 0 and currentTile.dstDir == "SOUTH": needsTeleport = True

        if needsTeleport:
            print(("Teleporting to grid: ", currentTile.dstGrid))
            palyer.grid = self.server.gridCache[currentTile.dstGrid]

            self.send_grid()
            player.posX = int(currentTile.dstX) * physics.tile_length
            player.posY = int(currentTile.dstY) * physics.tile_length

            for cli in list(self.server.game.clientsToPlayers.keys()):
                cli.send_move(player)
        else: 
          if player.grid.canStandOn(player.getTileX(moveX), player.getTileY(moveY)):
              player.moveRelative(moveX, moveY)
              player.walkState += .2

              if player.walkState > 2:
                  player.walkState = 0

              for cli in list(self.server.game.clientsToPlayers.keys()):
                  cli.send_move(player)

              destinationTile = player.grid.getTile(player.getTileX(), player.getTileY());

              if destinationTile.message is not None:
                  mesg = { 
                      "message": destinationTile.message
                  }

                  self.send("MESG", mesg)
          else:
              print("blocked")
              self.send("BLKD", {})

    def handle_halt(self, halt):
        self.server.halt()
