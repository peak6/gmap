package gmap

import (
	"encoding/hex"
	"flag"
	. "github.com/peak6/logger"
	"net"
	"time"
)

var autoDiscAddr = "239.20.0.3:9999"

func init() {
	flag.StringVar(&autoDiscAddr, "ada", autoDiscAddr, "AutoDiscover Address")
}

func StartAutoDiscover() error {
	mcaddr, err := net.ResolveUDPAddr("udp", autoDiscAddr)

	con, err := net.ListenUDP("udp", mcaddr)
	if err != nil {
		return err
	}
	go func() {
		buff := make([]byte, 2048)
		for {
			sz, addr, err := con.ReadFromUDP(buff)
			Linfo.Println("Got ping")
			if err != nil {
				Lerr.Println("Error reading message:", err)
			} else {
				Linfo.Printf("Got from: %s, bytes:\n%s", addr, hex.Dump(buff[0:sz]))
			}
		}
	}()
	go func() {
		c := time.Tick(1 * time.Second)
		for _ = range c {
			con.WriteToUDP([]byte{1, 2, 3}, mcaddr)
			Linfo.Println("Sent ping")
		}
		Linfo.Println("Exiting")
	}()
	return nil
}
