whole-test:
	go test ./...

name-test:
	go test ./... -run $(name)

build:
	go build -o bittorent cmd/main.go

info: build
	./bittorent info sample.torrent

peers: build
	./bittorent peers sample.torrent

handshake: build
	./bittorent handshake sample.torrent 178.62.85.20:51489

download_piece: build
	./bittorent download_piece -o /tmp/test-piece-0 sample.torrent 0

download_piece_test: build
	./bittorent download_piece -o /tmp/piece-9 test.torrent 3

download_piece_working: build
	./bittorent download_piece -o /tmp/piece-9-working working.torrent 9

download_file: build
	./bittorent download -o /tmp/test.txt working.torrent

codecrafters:
	rm bittorent || true
	codecrafters test
