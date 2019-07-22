class EventBus():                                                                  
  listeners = dict()
  commonData = None

  def __init__(self, commonData = None):
    self.commonData = commonData
                                                                                   
  def register(self, event, callback, data=None):
    if event not in self.listeners:                                                
      self.listeners[event] = list()                                               
                                                                                   
    self.listeners[event].append({
      "callback": callback,
      "data": data
    }) 
                                                                                   
  def fire(self, event, userData = None): 
    print("--- firing", event)                                                     
                                                                                   
    if event not in self.listeners:                                                
      return                                                                       
                                                                                   
    for trigger in self.listeners[event]:
      trigger["callback"](trigger['data'], self.commonData, userData)

  def fireStepOn(self, player):
    print("step on")

