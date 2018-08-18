import player
import logging

class game:
  # https://www.w3schools.com/colors/colors_crayola.asp
  availableColors = [
    0xff0000ff, # red 
    0x50bfe6ff, # blizzard blue
    0xF2C649ff, # maize 
    0x9C51B6ff, # purple plum
    0xAF6E4Dff, # brown sugar
    0x0048BAff, # absolute zero
    0xE936A7ff, # frostbite
    0xBFAFB2ff, # black shadows
    0x319177ff, # illuminating emerald
    0x757575ff, # sonic silver
  ]

  clientsToPlayers = {}
  lastPlayerId = 0
  server = None

  def __init__(self, server):
    self.server = server

  def registerClient(self, client):
    self.clientsToPlayers[client] = {
      "players": {}
    }

    client.grid = self.server.spawnGrid

  def registerPlayer(self, client, username):
    if self.getPlayerByUsername(username) != None:
      return # dont allow duplicate usernames

    self.lastPlayerId += 1

    plr = player.player(username, self.availableColors.pop(0))
    plr.id = self.lastPlayerId
    plr.grid = self.server.spawnGrid

    self.clientsToPlayers[client]["players"][username] = plr

    return plr

  def unregisterPlayer(self, client, username):
    plr = self.clientsToPlayers[client]["players"][username];
    del self.clientsToPlayers[client]["players"][username]

    self.availableColors.append(plr.color)

    for client in self.clientsToPlayers:
      client.send_player_quit(plr)

  def unregisterClient(self, client):
    players = list(self.clientsToPlayers[client]["players"].values())

    for otherClient in list(self.clientsToPlayers.keys()):
      for player in players:
        otherClient.send_player_quit(player)

    del self.clientsToPlayers[client]
      

  def getPlayerById(self, playerId):
    playerId = int(playerId)
    logging.debug("Getting player:" + str(playerId))

    for client in self.clientsToPlayers:
      for player in list(client["players"].values()):
        if player.id == playerId:
          return player

    return None

  def getPlayerByUsername(self, username):
    for client in list(self.clientsToPlayers.values()):
      for player in list(client["players"].values()):
        print(player)
        if username == player.username:
          return player

    return None
