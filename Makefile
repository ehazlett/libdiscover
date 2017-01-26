CGO_ENABLED=0
GOOS=linux
GOARCH=amd64

all: build

build:
	@cd example && go build -v .

clean:
	@rm -f example/example

.PHONY: build clean
