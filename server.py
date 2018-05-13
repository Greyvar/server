#!/bin/python

import SocketServer
import player
import map as grid
import tile
import threading
import game
import socket
import logging, logging.config
import client_interface
import signal
import os.path
from time import sleep

import cProfile

class server(SocketServer.ThreadingMixIn, SocketServer.TCPServer):
  allow_reuse_address = True
  gridCache = dict()
  run = True

  def halt(self):
    logging.debug("Server cleanly shutting down.");

    for client in self.game.clientsToPlayers:
      client.request.close()

    self.shutdown();

  def runTicker(self):
    while True:
      self.tick()
      sleep(.25)


  def tick(self):
    logging.debug("Server tick. " + str(len(self.game.clientsToPlayers)) + " clients.")

  def setup(self):
    self.spawnGrid = self.load_grid("dat/grids/1.1.grid")
    self.load_grid("dat/grids/1.2.grid")
    self.game = game.game(self)

  def load_grid(self, gridFilename):
    logging.debug("Loading grid: " + gridFilename)

    f = open(gridFilename)
    contents = f.read()
    f.close()

    ggrid = self.gridCache[os.path.basename(gridFilename)] = grid.map()

    for line in contents.split("\n"):
      if len(line) == 0: break

      if line[0] == '#': continue

      row, col, tex, rot, flipH, flipV, trv, tdstg, tdstx, tdsty, tdir, message, eol = line.split(",")
      row = int(row)
      col = int(col)
      rot = int(rot)
      trv = trv == "true"
      flipH = flipH == "true"
      flipV = flipV == "true"

      ggrid.setTile(row, col, tile.tile(tex, trv, rot, flipV, flipH, 
        dstGrid = tdstg,
        dstX = tdstx,
        dstY = tdsty,
        dstDir = tdir,
                message = message
      ))
      logging.debug("Parsed line in grid: " + line)

    return ggrid

def signal_handler(signal, frame):
  global srv, logging

  print # clear the signal written to terminal

  logging.info("Caught signal:" + str(signal))
  logging.info("Shutting down after signal.");

  srv.shutdown()

signal.signal(signal.SIGINT, signal_handler)

logging.config.fileConfig("etc/logging.conf")

srv = server(("localhost", 1337), client_interface.client_interface);
srv.setup();
server_thread = threading.Thread(target=srv.serve_forever)
server_thread.setDaemon(True)
server_thread.start();

server_tick_thread = threading.Thread(target=srv.runTicker)
server_tick_thread.setDaemon(True)
server_tick_thread.start()

while server_thread.isAlive():
  sleep(1)

