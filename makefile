whole-test:
	go test ./...

name-test:
	go test ./... -run $(name)

build:
	go build -o bittorent ./...

string: build
	./bittorent decode 5:hello
	./bittorent decode 10:strawberry

int: build
	./bittorent decode i345e

list: build
	./bittorent decode l5:helloi345ee
	./bittorent decode lli636e9:pineappleee
	./bittorent decode l10:strawberryi635ee
	./bittorent decode lli4eei5ee

info: build
	./bittorent info sample.torrent

peers: build
	./bittorent peers sample.torrent

codecrafters:
	rm bittorent || true
	codecrafters test
	codecrafters submit
