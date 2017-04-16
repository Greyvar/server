import player
import logging

class game:
	def __init__(self, server):
		self.clientsToPlayers = {}
		self.lastPlayerId = 0
		self.server = server

	def registerClient(self, client):
		self.clientsToPlayers[client] = None

	def registerPlayer(self, client, playerName):
		self.lastPlayerId += 1

		plr = player.player(playerName)
		plr.id = self.lastPlayerId
		plr.grid = self.server.spawnGrid

		self.clientsToPlayers[client] = plr

		client.send_player_you(plr);

		for client in self.clientsToPlayers.keys():
			client.send_player_join(plr)
			client.send_spawn(plr)

		return plr

	def unregisterPlayer(self, client):
		plr = self.clientsToPlayers[client];
		del self.clientsToPlayers[client]

		for client in self.clientsToPlayers:
			client.send_player_quit(plr)

	def getPlayerById(self, playerId):
		playerId = int(playerId)
		logging.debug("Getting player:" + str(playerId))

		for player in self.clientsToPlayers.values():
			if player.id == playerId:
				return player


	def playerMove(self, client):
		if (xOffset > 1 or yOffset > 1): return
		if (p.x + xOffset >= 16 - 1): return
		if (p.y + yOffset >= 16 - 1): return
		if (p.x + xOffset == 0): return;
		if (p.y + yOffset == 0): return;

		t = client.getGrid().getTile(p.x + xOffset, p.y + yOffset);

		if t.traversable:
			p.x += xOffset;
			p.y += yOffset;



