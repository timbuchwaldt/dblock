package dblockmain

import (
  "github.com/hpcloud/tail"
  "log"
  "sync"
  "time"
  "github.com/timbuchwaldt/dblock/blockstore"
  "github.com/timbuchwaldt/dblock/incidentstore"
  "github.com/timbuchwaldt/dblock/blocker"
  "flag"
  "net/http"
  "github.com/prometheus/client_golang/prometheus"
)

var addr = flag.String("listen-address", ":8080", "The address to listen for prometheus requests.")


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
  go incidentstore.IncidentStore(incidentChan, blockChan)
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

func init(){
  prometheus.MustRegister(incidentsCounter);
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
    incidentsCounter.Inc()
    // match against regexes here, if one matches, create "incident" struct
    c <- incidentstore.Incident{Filename: filename, Ip: "192.168.100.1", Time: time.Now(), Line: line.Text}
  }

}
