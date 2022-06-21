backupSource := ./cmd/snapshot
backupBinary := ./bin/bw-snapshot

build-all: clean
	GOOS=linux GOARCH=amd64 go build -o ./bin/amd64/bw-snapshot $(backupSource)
	GOOS=linux GOARCH=arm64 go build -o ./bin/arm64/bw-snapshot $(backupSource)
	GOOS=linux GOARCH=386 go build -o ./bin/386/bw-snapshot $(backupSource)

build: clean
	go build -o $(backupBinary) $(backupSource)

install: build
	install -m 0755 $(backupBinary) ${GOPATH}/bin/bw-snapshot

clean:
	rm -rf ./bin/ ./tmp/