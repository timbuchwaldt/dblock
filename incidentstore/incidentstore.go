package incidentstore

import (
  "time"
  "log"
  "github.com/timbuchwaldt/dblock/blockstore"

)

type Incident struct {
  Filename  string
  Ip        string
  Time      time.Time
  Line      string
}

func IncidentStore(incidentChan chan Incident, blockChan chan blockstore.Block) {
  ticker := time.NewTicker(time.Second * 5)
  var incidents []Incident
  var msg Incident
  for {
    select {
    case msg = <-incidentChan:
        log.Println("[incidentStore] " + msg.Ip)
        incidents = append(incidents, msg)
      case <-ticker.C:
        // cleanup
        var newIncidents [] Incident
        for _, incident := range incidents {
          if time.Since(incident.Time) < 5*time.Second{
            newIncidents = append(newIncidents, incident)
          }
        }
        incidents = newIncidents

        // count violations per IP
        var buckets = make(map[string]int)
        for _, incident := range incidents {
          buckets[incident.Ip] += 1
        }
        log.Println(buckets)
        for ip, violations := range buckets{
          if violations > 5 {
            log.Println("ALARM ALARM")
            blockChan <- blockstore.Block{Ip: ip, Time: time.Now(), External: false}

          }
        }

    }
  }
}
