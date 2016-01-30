package dblockmain

import (
	"github.com/hpcloud/tail"
	"log"
	"sync"
  "time"
  "github.com/timbuchwaldt/dblock/blockstore"
  "github.com/timbuchwaldt/dblock/incidentstore"
	"github.com/timbuchwaldt/dblock/blocker"
)



func Main() {
	// read Config

	// compile regex
	// start sync
	// start storage
	// start controller
	// start follower + analyzer

	var wg sync.WaitGroup
	wg.Add(1)

	log.Println("Hello world")
	incidentChan := make(chan incidentstore.Incident, 100)
  blockChan := make(chan blockstore.Block, 100)
	blockControlChan := make(chan blocker.ControlMsg, 100)

	go blocker.Blocker(blockControlChan)
  go blockstore.BlockStore(blockChan, blockControlChan)
  go incidentstore.IncidentStore(incidentChan, blockChan)
	go follow_and_analyze("foo", incidentChan)
	go follow_and_analyze("bar", incidentChan)


	wg.Wait()

}

func follow_and_analyze(filename string, c chan incidentstore.Incident) {
	t, err := tail.TailFile(filename,
		tail.Config{
			Follow:   true,
			ReOpen:   true,
			Location: &tail.SeekInfo{Offset: 0, Whence: 2}})

	if err != nil {
		log.Fatal(err)
	}

	for line := range t.Lines {

		// match against regexes here, if one matches, create "incident" struct
		c <- incidentstore.Incident{Filename: filename, Ip: "192.168.100.1", Time: time.Now(), Line: line.Text}
	}

}
