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

func Blocker(controlChan chan ControlMsg, stringWhitelist []string) {

	var whitelist []net.IPNet
	for _, string_entry := range stringWhitelist {
		_, network, err := net.ParseCIDR(string_entry)
		if err != nil {
			log.Fatal(err)
		}
		whitelist = append(whitelist, *network)
	}

	executeCommand("create dblock hash:ip maxelem 1048576 -exist")
	executeCommand("flush dblock")
	executeCommand("create dblock6 hash:ip maxelem 1048576 inet6 -exist")
	executeCommand("flush dblock6 ")

	for msg := range controlChan {
		whitelisted := false
		for _, entry := range whitelist {
			if entry.Contains(msg.Ip) {
				whitelisted = true
			}
		}
		if whitelisted {
			log.Println("[sync]\tIP is part of whitelisted networks")
		} else {
			if msg.Block {
				if msg.Ip.To4() != nil {
					// this is an ipv4 address
					log.Println("[blocker]\tBlocking ip " + msg.Ip.String())
					executeCommand("add dblock " + msg.Ip.String() + " -exist")
				} else {
					// this is an ipv6 address
					log.Println("[blocker]\tBlocking ipv6 " + msg.Ip.String())
					executeCommand("add dblock6 " + msg.Ip.String() + " -exist")
				}
			} else {
				if msg.Ip.To4() != nil {
					// this is an ipv4 address
					log.Println("[blocker]\tUnblocking ip " + msg.Ip.String())
					executeCommand("del dblock " + msg.Ip.String() + " -exist")
				} else {
					// this is an ipv6 address
					log.Println("[blocker]\tUnblocking ipv6 " + msg.Ip.String())
					executeCommand("del dblock6 " + msg.Ip.String() + " -exist")
				}
			}
		}
	}
}

func executeCommand(arguments string) {
	err := exec.Command("/sbin/ipset", arguments).Run()
	if err != nil {
		log.Fatal(err)
	}
}
