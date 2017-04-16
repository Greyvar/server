def consumeNumber(self, line):
	if line.find(",") == -1:
		return ["", int(line)]

	number, remainder = line.split(",", 1)
	
	return [remainder, int(number)]

