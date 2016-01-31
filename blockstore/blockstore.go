package blockstore

import (
	"github.com/timbuchwaldt/dblock/blocker"
	"log"
	"net"
	"time"
)

type Block struct {
	Ip       net.IP
	Time     time.Time
	External bool
}

func BlockStore(blockChan chan Block, blockControlChan chan blocker.ControlMsg) {
	ticker := time.NewTicker(time.Second * 5)
	var blocks []Block
	for {
		select {
		case msg := <-blockChan:
			// check if we already blocked this IP
			found := false
			for _, element := range blocks {
				if element.Ip.Equal(msg.Ip) {
					found = true
					break
				}
			}
			if !found {
				log.Println("[blockstore] Blocking " + msg.Ip.String())
				blockControlChan <- blocker.ControlMsg{Ip: msg.Ip, Block: true}
				blocks = append(blocks, msg)
			}
		case <-ticker.C:
			// clean out blocks that are older than x
			var newBlocks []Block
			for _, block := range blocks {
				log.Println(block)
				if time.Since(block.Time) < 30*time.Second {
					newBlocks = append(newBlocks, block)
				} else {
					log.Println("unblocking: " + block.Ip.String())
					blockControlChan <- blocker.ControlMsg{Ip: block.Ip, Block: false}
				}
			}
			blocks = newBlocks
		}
	}
}
