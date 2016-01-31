package blocker

import (
	"log"
	"net"
	"os/exec"
)

type ControlMsg struct {
	Ip    net.IP
	Block bool
}

func Blocker(controlChan chan ControlMsg) {

	var whitelist [2]net.IPNet
	_, local, _ := net.ParseCIDR("127.0.0.1/8")
	whitelist[0] = *local

	exec.Command("ipset", "create dblock hash:ip maxelem 1048576 -exist").Run()        // create v4 set, -exist allows creation even if it exists
	exec.Command("ipset", "flush dblock").Run()                                        // flush old rules
	exec.Command("ipset", "create dblock6 hash:ip maxelem 1048576 inet6 -exist").Run() // create v6 set, -exist allows creation even if it exists
	exec.Command("ipset", "flush dblock6 ").Run()                                      // flush old rules

	for msg := range controlChan {
		if local.Contains(msg.Ip) {
			log.Println("IP is part of whitelisted networks")
		} else {
			if msg.Block {
				if msg.Ip.To4() != nil {
					// this is an ipv4 address
					log.Println("[blocker]\tBlocking ip " + msg.Ip.String())
					exec.Command("ipset", "add dblock "+msg.Ip.String()+" -exist").Run()
				} else {
					// this is an ipv6 address
					log.Println("[blocker]\tBlocking ipv6 " + msg.Ip.String())
					exec.Command("ipset", "add dblock6 "+msg.Ip.String()+" -exist").Run()
				}
			} else {
				if msg.Ip.To4() != nil {
					// this is an ipv4 address
					log.Println("[blocker]\tUnblocking ip " + msg.Ip.String())
					exec.Command("ipset", "del dblock "+msg.Ip.String()+" -exist").Run()
				} else {
					// this is an ipv6 address
					log.Println("[blocker]\tUnblocking ipv6 " + msg.Ip.String())
					exec.Command("ipset", "del dblock6 "+msg.Ip.String()+" -exist").Run()
				}
			}
		}
	}
}
