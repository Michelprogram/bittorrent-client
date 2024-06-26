package torrent

import (
	"fmt"
	"testing"
)

func BenchmarkDownloadLow(b *testing.B) {

	client, err := NewBittorrent("../../sample.torrent")

	if err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		client.DownloadLow(fmt.Sprintf("test-%d.txt", i))
	}
}

func BenchmarkDownloadFast(b *testing.B) {

	client, err := NewBittorrent("../../sample.torrent")

	if err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		client.DownloadFast(fmt.Sprintf("test-%d.txt", i))
	}
}
