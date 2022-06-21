backupSource := ./cmd/snapshot
backupBinary := ./bin/bw-snapshot

build: clean
	go build -o $(backupBinary) $(backupSource)

install: build
	install -m 0755 $(backupBinary) ${GOPATH}/bin/bw-snapshot

clean:
	rm -rf ./bin/ ./tmp/