import random

class player:
	def __init__(self, nickname):
		self.x = 3;
		self.y = 3;
		self.nickname = nickname
		self.id = None
		self.skin = random.choice(["Purple"])
		self.grid = None

	def getGrid(self):
		return self.grid

	def getCurrentTile(self):
		return self.grid.getTile(self.x, self.y)
