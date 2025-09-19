.PHONY: all install clean

all: clean install

install:
	@echo "Installing Go program..."
	go install -ldflags "-X main.BuildTime=$(shell date '+%Y-%m-%dT%H:%M:%S%Z')"
	@echo "Installation complete."

clean:
	@echo "Cleaning up..."
	go clean -x -i ./... 
	@echo "Cleanup complete."
