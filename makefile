string:
	go run cmd/mybittorrent/main.go decode 5:hello
	go run cmd/mybittorrent/main.go decode 10:strawberry

int:
	go run cmd/mybittorrent/main.go decode i345e

list:
	go run cmd/mybittorrent/main.go decode l5:helloi345ee
	go run cmd/mybittorrent/main.go decode lli636e9:pineappleee
	go run cmd/mybittorrent/main.go decode l10:strawberryi635ee
	go run cmd/mybittorrent/main.go decode lli4eei5ee
