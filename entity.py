class Entity:
  id = 0
  definition = None
  state = None

  def __init__(self, definition, state):
    self.definition = definition
    self.state = state
