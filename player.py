import random

class player:
  def __init__(self, nickname):
    self.x = 12;
    self.y = 7;
    self.nickname = nickname
    self.id = None
    self.skin = random.choice(["Purple"])
    self.walkState = 0
    self.grid = None

  def getGrid(self):
    return self.grid

  def getCurrentTile(self):
    return self.grid.getTile(round(self.x), round(self.y))
