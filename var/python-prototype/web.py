import logging
import cherrypy
import yaml

class Api:
  game = None

  def __init__(self, game):
    self.game = game

  @cherrypy.expose
  def index(self, *args, **kwargs):
    payload = {
      "title": "greyvar-server",
      "players": len(self.game.clientsToPlayers)
    }

    return self.toYaml(payload);
    
  def toYaml(self, payload):
    cherrypy.response.headers["Content-Type"] = "text/plain"

    return yaml.dump(payload, default_flow_style = False)

def http_error_handler(status, message, traceback, version):
  return json.dumps({
    "httpStatus": status,
    "type": "httpError",
    "message": message
  });


def error_handler():
  cherrypy.response.status = 500;
  cherrypy.response.headers['Content-Type'] = 'text/plain'

  exceptionInfo = exc_info()
  excType = exceptionInfo[0]
  exception = exceptionInfo[1]

  cherrypy.response.body = "\nUnhandled exception.\n" + "Message: " + exception.message + "\n" + "Type: " + str(excType.__name__)

  print(exceptionInfo)

def start(game):
  api = Api(game);

  cherrypy.config.update({ 
    'server.socket_host': "0.0.0.0",                                         
    'server.socket_port': 1380,                                                 
    'tools.sessions.on': False,                                                       
    'tools.sessions.storage_type': 'ram',
    'tools.sessions.timeout': 0,                                  
    'request.error_response': error_handler,                                         
    'request.error_page': {'default': http_error_handler }                           
  })                                

  appConfig = {}
                                                                                     
  #cherrypy.process.plugins.Daemonizer(cherrypy.engine).subscribe()
  cherrypy.quickstart(api, "/", {"/": appConfig})

def stop():
  logging.info("Shutting down web")
  cherrypy.engine.exit()

