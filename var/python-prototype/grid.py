import tile
import logging

class Grid:
  height = 0
  width = 0

  def __init__(self, width, height):
    # deliberately flip these, because, err, bug.
    self.height = width 
    self.width = height

    print("width: " + str(self.width) + ", height: " + str(self.height))
 
    self.rowIterator = range(0, self.height)
    self.colIterator = range(0, self.width)

    self.initGrid();

  def __repr__(self):
    return "grid{}"

  def initGrid(self):
    self.tiles = {}
    self.entities = {}

    for row in self.rowIterator:
      self.tiles[row] = {}
      self.entities[row] = {}

      for col in self.colIterator:
        self.tiles[row][col] = None
        self.entities[row][col] = None

  def getTile(self, x, y):
    tileFound = True

    if x - 1 > len(self.tiles) or y - 1 > len(self.tiles[x]):
      tileFound = False
    elif self.tiles[x][y] == None:
      tileFound = False

    if not tileFound: 
      t = tile.tile();
      t.tex = "grass.png"
      return t
    else:
      return self.tiles[x][y]

  def getEntity(self, row, col):
    try:
      if self.entities[row, col] != None:
        return self.entities[row][col]
    except:
      return None
    
  def setTile(self, row, col, tile):
    if row >= self.height:
        print(row, col, str(tile))
        logging.info("row is out of bounds" + str(row))

    if col >= self.width:
        print(row, col, str(tile))
        logging.info("col is out of bounds" + str(col))

    self.tiles[row][col] = tile;

  def setEntity(self, row, col, ent):
    self.entities[row][col] = ent

  def allTiles(self):
    for row in self.rowIterator:
      for col in self.colIterator:
        yield [row, col, self.getTile(row, col)]

  def allEntities(self):
    for row in self.rowIterator:
      for col in self.colIterator:
        yield [row, col, self.entities[row][col]]

  def canStandOn(self, x, y):
    if y >= len(self.tiles) or y < 0:
      return False;

    if x >= len(self.tiles[0]) or x < 0:
      return False;

    return self.getTile(x, y).traversable;

