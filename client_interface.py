import SocketServer
import player
import map
import tile
import threading
import game
import socket
import logging, logging.config

class client_interface(SocketServer.StreamRequestHandler):
    def setup(self):
        self.alive = True
        self.server.game.registerClient(self);

    def handle(self):
        logging.debug("New connection.");

        while self.alive:
            try:
                line = self.request.recv(1024); 
            except socket.error as e:
                logging.debug("socket exception: " + str(e))
                break;

            line = line.decode("utf-8")

            if not line: break;
            if not line.find("\n"): continue;
    

            cmd = line[0:4]

            if cmd == "HELO":
                self.handle_helo(line)
            elif cmd == "QUIT":
                return;
            elif cmd == "HALT":
                self.handle_halt(line);
            elif cmd == "MOVR":
                self.handle_movr(line);
            else:
                logging.debug("Unknown command from client: " + str(line));

        self.server.game.unregisterPlayer(self)

    def finish(self):
        self.alive = False

    def send_player_you(self, player):
        self.send("PLRU", player.id, player.nickname, player.skin)

    def send_player_join(self, player):
        logging.debug("sending new player")
        self.send("PLRJ", player.id, player.nickname, player.skin);
        self.send_move(player)

    def send_player_quit(self, player):
        if player != None:
            self.send("PLRQ", player.id)

    def send_spawn(self, player):
        self.send("SPWN", player.x, player.y);

    def send_move(self, player):
        self.send("MOVE", player.id, player.x, player.y)

    def send_grid(self):
        for row, col, tile in self.localPlayer.grid.allTiles():
            self.send("TILE", row, col, tile.tex, tile.rot, tile.flipV, tile.flipH);

    def send(self, command, *args):
        message = command.strip().upper() + " " + ",".join(["%s" % el for el in args]) + "\n"
        self.request.send(message.encode("utf-8"))

    def handle_helo(self, line):
        self.localPlayer = self.server.game.registerPlayer(self, line[5:15]);
        self.send_grid();

        for plr in self.server.game.clientsToPlayers.values():
            self.send_player_join(plr)

    def handle_movr(self, line):
        line = line[5:999]
        line, playerId = self.consumeNumber(line)
        line, moveX = self.consumeNumber(line)
        line, moveY = self.consumeNumber(line)

        moveX = int(moveX)
        moveY = int(moveY)

        currentTile = self.localPlayer.getCurrentTile()
        needsTeleport = False

        print moveX, moveY, currentTile.dstDir, currentTile.dstGrid

        if currentTile.dstGrid != "":
            if moveX > 0 and currentTile.dstDir == "EAST": needsTeleport = True
            if moveX < 0 and currentTile.dstDir == "WEST": needsTeleport = True
            if moveY > 0 and currentTile.dstDir == "NORTH": needsTeleport = True
            if moveY < 0 and currentTile.dstDir == "SOUTH": needsTeleport = True

        destinationTile = self.localPlayer.grid.getTile(self.localPlayer.x + moveX, self.localPlayer.y);

        if needsTeleport:
            print "Moving to grid: ", currentTile.dstGrid
            self.localPlayer.grid = self.server.gridCache[currentTile.dstGrid]

            self.send_grid()
            self.localPlayer.x = int(currentTile.dstX)
            self.localPlayer.y = int(currentTile.dstY)

            for cli in self.server.game.clientsToPlayers.keys():
                cli.send_move(self.localPlayer)


        elif self.localPlayer.grid.canStandOn(self.localPlayer.x + moveX, self.localPlayer.y + moveY):
            self.localPlayer.x += int(moveX)
            self.localPlayer.y += int(moveY)

            for cli in self.server.game.clientsToPlayers.keys():
                cli.send_move(self.localPlayer)

        if destinationTile.message is not None:
            self.send("MESG", destinationTile.message)

    def handle_halt(self, line):
        line = line[5:999];

        self.server.halt()

    def consumeNumber(self, line):
        if line.find(",") == -1:
            return ["", int(line)]

        number, remainder = line.split(",", 1)
        
        return [remainder, int(number)]
