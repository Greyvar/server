import random
import time
import physics
import math

class player:
  def __init__(self, nickname):
    self.posX = 32
    self.posY = 32
    self.nickname = nickname
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
    return int(math.ceil((pos + 8 + offset) / 16.0))

  def getGrid(self):
    return self.grid

  def getCurrentTile(self):
    return self.grid.getTile(self.getTileX(), self.getTileY())
