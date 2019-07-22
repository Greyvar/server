import random
import time
import physics
import math
import binascii

class player:
  def __init__(self, username, color):
    self.posX = 4 * physics.tile_length
    self.posY = 4 * physics.tile_length
    self.username = username
    self.color = color
    self.id = None
    self.skin = random.choice(["Purple"])
    self.walkState = 0
    self.grid = None
    self.timeOfLastMove = time.time()

  def getTileX(self, offset = 0):
    return self.getTilePos(self.posX, offset);

  def getTileY(self, offset = 0):
    return self.getTilePos(self.posY, offset);

  def getTilePos(self, pos, offset = 0):
    return int(math.ceil((pos + (offset * physics.player_move_speed)) / physics.player_move_speed))

  def moveRelative(self, x, y):
    self.posX += x * physics.player_move_speed
    self.posY += y * physics.player_move_speed

  def getGrid(self):
    return self.grid

  def getCurrentTile(self):
    return self.grid.getTile(self.getTileX(), self.getTileY())
