package dblockmain

import (
	"flag"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/timbuchwaldt/dblock/blocker"
	"github.com/timbuchwaldt/dblock/blockstore"
	"github.com/timbuchwaldt/dblock/incidentstore"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"
)

var (
	addr          = flag.String("listen-address", ":8080", "The address to listen for prometheus requests.")
	timebucket    = flag.Duration("incident-bucket", 5*time.Minute, "The number of seconds of incidents we compare.")
	max_incidents = flag.Int("max-incidents", 5, "The number incidents allowed per incident-bucket.")
)

func Main() {
	flag.Parse()
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(*addr, nil)

	var wg sync.WaitGroup
	wg.Add(1)

	log.Println("Hello world")
	incidentChan := make(chan incidentstore.Incident, 100)
	blockChan := make(chan blockstore.Block, 100)
	blockControlChan := make(chan blocker.ControlMsg, 100)

	go blocker.Blocker(blockControlChan)
	go blockstore.BlockStore(blockChan, blockControlChan)
	go incidentstore.IncidentStore(incidentChan, blockChan, *timebucket, *max_incidents)
	go follow_and_analyze("foo", incidentChan)
	go follow_and_analyze("bar", incidentChan)

	wg.Wait()

}

var (
	incidentsCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "incidents_total",
		Help: "Number of total dblock incidents",
	})
)

func init() {
	prometheus.MustRegister(incidentsCounter)
}

func follow_and_analyze(filename string, c chan incidentstore.Incident) {
	t, err := tail.TailFile(filename,
		tail.Config{
			Follow:   true,                                  // actually follow the logs
			ReOpen:   true,                                  // allow logs to be rotated
			Location: &tail.SeekInfo{Offset: 0, Whence: 2}}) // seek to end of file

	if err != nil {
		log.Fatal(err)
	}

	regex := regexp.MustCompile("Accepted publickey for .+ from (?P<ip>.+) port .*")

	for line := range t.Lines {
		incidentsCounter.Inc()
		result := regex.FindStringSubmatch(line.Text)
		if result != nil {
			log.Println(result[1])
			c <- incidentstore.Incident{Filename: filename, Ip: "192.168.100.1", Time: time.Now(), Line: line.Text}
		}
		// match against regexes here, if one matches, create "incident" struct
	}

}
