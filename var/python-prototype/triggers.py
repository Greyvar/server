def onLoaded(one, two, three):
    print("onLoadedddddd trigger", one, two)

def stepOn(common, params, player):
    print("step on", common, params, player)

class ConditionPlayerWalkInto():
    def __init__(self, x, y):
        self.x = x
        self.y = y

class ActionMessage():
    message = "unknown"

class Trigger():
    conditions = [
        ConditionPlayerWalkInto(3, 3)
    ]

    actions = [
        ActionMessage()
    ]


