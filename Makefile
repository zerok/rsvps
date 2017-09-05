all: rsvps

rsvps: $(shell find . -name '*.go')
	cd cmd/rsvps && go build -o ../../rsvps

linux:
	cd cmd/rsvps && GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o ../../rsvps

clean:
	rm -f rsvps
