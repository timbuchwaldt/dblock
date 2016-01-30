package blockstore

import (
  "time"
  "log"
  "github.com/timbuchwaldt/dblock/blocker"
)

type Block struct {
  Ip        string
  Time      time.Time
  External  bool
}

func BlockStore(blockChan chan Block, blockControlChan chan blocker.ControlMsg) {
  ticker := time.NewTicker(time.Second * 5)
  var blocks []Block
  for {
    select {
    case msg := <-blockChan:
        log.Println("[blockStore] " + msg.Ip)
        blockControlChan <- blocker.ControlMsg{Ip: msg.Ip, Block: true}
        blocks = append(blocks, msg)
      case <-ticker.C:
        // clean out blocks that are older than x
        var newBlocks [] Block
        for _, block := range blocks {
          log.Println(block)
          if time.Since(block.Time) < 5*time.Second{
            newBlocks = append(newBlocks, block)
          }else{
            log.Println("unblocking: " + block.Ip)
            blockControlChan <- blocker.ControlMsg{Ip: block.Ip, Block: false}
          }
        }
        blocks = newBlocks
    }
  }
}
