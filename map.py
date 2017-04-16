import tile

class map:
	def __init__(self):
		self.initMap();

	def initMap(self):
		self.tiles = {}

		for row in range(0, 16):
			self.tiles[row] = {}

			for col in range(0, 16):
				self.tiles[row][col] = None

	def getTile(self, x, y):
		if self.tiles[x][y] == None:
			t = tile.tile();
			t.tex = "grass.png"
			return t
		else:
			return self.tiles[x][y]
		
	def setTile(self, row, col, tile):
		self.tiles[row][col] = tile;

	def allTiles(self):
		for row in range(0, 16):
			for col in range(0, 16):
				yield [row, col, self.getTile(row, col)]

	def canStandOn(self, x, y):
		if y >= len(self.tiles) or y < 0:
			return False;

		if x >= len(self.tiles[0]) or x < 0:
			return False;

		return self.getTile(x, y).traversable;
