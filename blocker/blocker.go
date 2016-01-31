package blocker

import (
	"log"
	"net"
)

type ControlMsg struct {
	Ip    net.IP
	Block bool
}

func Blocker(controlChan chan ControlMsg) {
	log.Print("ipset create dblock hash:ip maxelem 1048576 -exist")
	log.Print("ipset create dblock6 hash:ip maxelem 1048576 inet6 -exist")

	for msg := range controlChan {
		if msg.Block {
			if msg.Ip.To4() != nil {
				// this is an ipv4 address
				log.Println("Blocking ip " + msg.Ip.String())
				log.Print("ipset add dblock " + msg.Ip.String() + " -exist")
			} else {
				// this is an ipv6 address
				log.Println("Blocking ip " + msg.Ip.String())
				log.Print("ipset add dblock6 " + msg.Ip.String() + " -exist")
			}
		} else {
			if msg.Ip.To4() != nil {
				// this is an ipv4 address
				log.Println("Blocking ip " + msg.Ip.String())
				log.Print("ipset del dblock " + msg.Ip.String() + " -exist")
			} else {
				// this is an ipv6 address
				log.Println("Blocking ip " + msg.Ip.String())
				log.Print("ipset del dblock6 " + msg.Ip.String() + " -exist")
			}
		}
	}
}
