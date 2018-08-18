class Tile:
    def __init__(self, tex = None, traversable = True, rot = 0, flipV = False, flipH = False, dstGrid = None, dstX = 0, dstY = 0, dstDir = None, message = None):
        self.tex = tex
        self.traversable = traversable
        self.rot = rot
        self.flipV = flipV
        self.flipH = flipH

        self.dstGrid = dstGrid
        self.dstX = dstX
        self.dstY = dstY
        self.dstDir = dstDir

        if message == "" or message == "null":
            self.message = None
        else:
            self.message = message
