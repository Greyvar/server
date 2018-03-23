import SocketServer
import player
import map
import tile
import threading
import game
import socket
import logging, logging.config
import yaml
import math

ETB = chr(0x17)

class client_interface(SocketServer.StreamRequestHandler):
    def setup(self):
        self.alive = True
        self.server.game.registerClient(self);

    def handle(self):
        logging.debug("New connection.");

        welc = {
            "serverVersion": "greyvar-devel"
        }

        self.send("WELC", welc)

        chunkBuf = ""

        while self.alive:
            try:
                chunk = self.request.recv(1024); 
            except socket.error as e:
                logging.debug("socket exception: " + str(e))
                break;

            if not chunk: break;

            print "recv", chunk

            if chunk.find(ETB):
                chunkEnd, nextChunkStart = chunk.split(ETB)
                chunkBuf += chunkEnd

                self.parse_chunk(chunkBuf)
                chunkBuf = nextChunkStart

        self.server.game.unregisterPlayer(self)

    def parse_chunk(self, chunk):
        logging.debug("parse chunk: " + str(chunk));

        req = yaml.load(chunk)
        cmd = req["command"]

        if cmd == "HELO":
            self.handle_helo(req)
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

        return dict(filter(lambda i:i[0] in args, d.items()))

    def finish(self):
        self.alive = False

    def send_player_you(self, player):
        plru = self.subdict(player, "id", "nickname", "skin");

        self.send("PLRU", plru)

    def send_player_join(self, player):
        logging.debug("sending new player")

        plrj = self.subdict(player, "id", "nickname", "skin")

        self.send("PLRJ", plrj);
        self.send_move(player)

    def send_player_quit(self, player):
        if player != None:
            self.send("PLRQ", self.subdict(player, "id"))

    def send_spawn(self, player):
        spwn = self.subdict(player, "x", "y")

        self.send("SPWN", spwn);

    def send_move(self, player):
        move = {
            "posX": player.x,
            "posY": player.y,
            "playerId": player.id,
            "walkState": int(math.floor(player.walkState))
        }

        self.send("MOVE", move)


    def send_grid(self):
        for row, col, tile in self.localPlayer.grid.allTiles():

            tile = {
                "row": row,
                "col": col,
                "tex": tile.tex,
                "rot": tile.rot,
                "flipV": tile.flipV,
                "flipH": tile.flipH
            }

            self.send("TILE", tile);

    def send(self, command, payload):
        payload["apiVersion"] = "v1"
        payload["command"] = command.strip().upper()
        payload = yaml.dump(payload)
        
        print "send", payload

        message = payload + "\n"
        self.request.send(message.encode("utf-8") + ETB)

    def handle_helo(self, helo):
        self.localPlayer = self.server.game.registerPlayer(self, helo['username']);
        self.send_grid();

        for plr in self.server.game.clientsToPlayers.values():
            self.send_player_join(plr)

    def handle_movr(self, movr):
        moveX = movr['x']
        moveY = movr['y']

        currentTile = self.localPlayer.getCurrentTile()
        needsTeleport = False

        print moveX, moveY, currentTile.dstDir, currentTile.dstGrid, self.localPlayer.grid 

        if currentTile.dstGrid != "":
            logging.debug("this tile has a destination: " + currentTile.dstDir)
            if moveX > 0 and currentTile.dstDir == "EAST": needsTeleport = True
            if moveX < 0 and currentTile.dstDir == "WEST": needsTeleport = True
            if moveY > 0 and currentTile.dstDir == "NORTH": needsTeleport = True
            if moveY < 0 and currentTile.dstDir == "SOUTH": needsTeleport = True

        if needsTeleport:
            print "Teleporting to grid: ", currentTile.dstGrid
            self.localPlayer.grid = self.server.gridCache[currentTile.dstGrid]

            self.send_grid()
            self.localPlayer.x = int(currentTile.dstX)
            self.localPlayer.y = int(currentTile.dstY)

            for cli in self.server.game.clientsToPlayers.keys():
                cli.send_move(self.localPlayer)
        elif self.localPlayer.grid.canStandOn(round(self.localPlayer.x + moveX), round(self.localPlayer.y + moveY)):
            self.localPlayer.x += moveX
            self.localPlayer.y += moveY
            self.localPlayer.walkState += .2

            if self.localPlayer.walkState > 2:
                self.localPlayer.walkState = 0

            for cli in self.server.game.clientsToPlayers.keys():
                cli.send_move(self.localPlayer)

            destinationTile = self.localPlayer.grid.getTile(round(self.localPlayer.x), round(self.localPlayer.y));

            if destinationTile.message is not None:
                mesg = { 
                    "message": destinationTile.message
                }

                self.send("MESG", mesg)

    def handle_halt(self, halt):
        self.server.halt()
