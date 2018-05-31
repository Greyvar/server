#!/bin/python

import SocketServer
import player
import threading
import game
import socket
import logging, logging.config
import client_interface
import signal
import os.path
from time import sleep
from grid import Grid
from tile import Tile
import yaml

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
    self.load_world("isleOfStarting_dev")

    self.game = game.game(self)

  def load_world(self, worldName):
    try: 
      worldFile = open("dat/worlds/" + worldName + "/world.yml", "r")
      worldDef = yaml.load(worldFile.read())
      worldFile.close()

      gridName = worldDef["spawnGrid"]

      self.spawnGrid = self.load_grid("dat/worlds/" + worldName + "/grids/" + gridName)

    except Exception as e: 
      print "Failed to load world", str(e)

  def load_grid(self, gridFilename):
    logging.debug("Loading grid: " + gridFilename)

    gridFile = open(gridFilename)
    yamlContents = yaml.load(gridFile.read())
    gridFile.close()

    ggrid = self.gridCache[os.path.basename(gridFilename)] = Grid()

    for tile in yamlContents["tiles"]:
      t = Tile(tile['texture'], tile['traversable'], tile['rot'], tile['flipV'], tile['flipH'])

      ggrid.setTile(tile['x'], tile['y'], t)

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

