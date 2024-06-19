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
	go run cmd/mybittorrent/main.go decode i345e

list:
	go run cmd/mybittorrent/main.go decode l5:helloi345ee
	go run cmd/mybittorrent/main.go decode lli636e9:pineappleee
	go run cmd/mybittorrent/main.go decode l10:strawberryi635ee
	go run cmd/mybittorrent/main.go decode lli4eei5ee
