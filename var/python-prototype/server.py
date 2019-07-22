#!/bin/python3

import sys
import socketserver
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
from entity import Entity
import yaml
import web
from EventBus import EventBus
from triggers import *

import cProfile

class server(socketserver.ThreadingMixIn, socketserver.TCPServer):
  allow_reuse_address = True
  gridCache = dict()
  run = True

  triggerEventBus = None

  entdefs = dict()

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
    self.game = game.game(self)

    cd = {
        "server": self
    }

    self.triggerEventBus = EventBus(cd)
#    self.triggerEventBus.register("world_loaded", triggers.onLoaded)
#    self.triggerEventBus.register("stepOn", triggers.stepOn)

    self.load_entdefs()
    self.load_world("isleOfConcept_dev")

  def load_entdefs(self):
    try: 
      for f in os.listdir("dat/entdefs/"):
        if f.endswith(".yml"):
          entdef = self.load_entdef_file(f)

          self.entdefs[entdef['title']] = entdef
    except Exception as e:
      logging.error("Excepting loading entdefs", type(e), str(e), e)
      
  def load_entdef_file(self, f):
    handle = open("dat/entdefs/" + f, 'r')
    entdef = yaml.load(handle.read())
    handle.close()

    logging.info("Loaded entdef " + entdef['title'] + " from " + f)

    return entdef

  def load_world(self, worldName):
    worldFile = open("dat/worlds/" + worldName + "/world.yml", "r")
    worldDef = yaml.load(worldFile.read())
    worldFile.close()

    gridName = worldDef["spawnGrid"]

    self.spawnGrid = self.load_grid("dat/worlds/" + worldName + "/grids/" + gridName)

    for trigger in worldDef['triggers']:
      self.load_world_trigger(trigger)

    self.triggerEventBus.fire("world_loaded")

  def load_world_trigger(self, trigger):
    testTrigger = Trigger()
    testTrigger.conditions.append(ConditionPlayerWalkInto(3, 3))
    testTrigger.actions.append(ActionMessage())


  def load_grid(self, gridFilename):
    logging.debug("Loading grid: " + gridFilename)

    gridFile = open(gridFilename)
    yamlContents = yaml.load(gridFile.read())
    gridFile.close()

    ggrid = self.gridCache[os.path.basename(gridFilename)] = Grid(int(yamlContents['width']), int(yamlContents['height']))

    logging.info("loading tiles")
    for tile in yamlContents["tiles"]:
      print(tile)
      t = Tile(tile['texture'], tile['traversable'], tile['rot'], tile['flipV'], tile['flipH'])

      ggrid.setTile(tile['x'], tile['y'], t)

    logging.info("loading entities");
    for ent in yamlContents["entities"]:
      logging.info("Loaded entity")

      if ent['definition'] not in self.entdefs:
        logging.warn("Found entity instance with unregistered entdef: " + ent['definition'])
      else:
        entdef = self.entdefs[ent['definition']]

        e = Entity(ent["definition"], entdef['initialState'])
        e.id = int(ent["id"])

        ggrid.setEntity(int(ent["x"]), int(ent["y"]), e)

    logging.info("Loaded all entities")
    return ggrid

def signal_handler(signal, frame):
  global srv, logging

  print() # clear the signal written to terminal

  logging.info("Caught signal:" + str(signal))
  logging.info("Shutting down after signal.");

  web.stop()

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

web_thread = threading.Thread(target=web.start, args=[srv.game])
web_thread.setDaemon(True)
web_thread.start()

while server_thread.isAlive():
  sleep(1)

