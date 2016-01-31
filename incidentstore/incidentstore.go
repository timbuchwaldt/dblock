package incidentstore

import (
	"github.com/timbuchwaldt/dblock/blockstore"
	"log"
	"net"
	"time"
)

type Incident struct {
	Filename string
	Ip       net.IP
	Time     time.Time
	Line     string
}

func IncidentStore(incidentChan chan Incident, blockChan chan blockstore.Block, timebucket time.Duration, max_incidents int) {
	ticker := time.NewTicker(time.Second * 5)
	var incidents []Incident
	for {
		select {
		case msg := <-incidentChan:
			incidents = append(incidents, msg)
		case <-ticker.C:
			// cleanup
			var newIncidents []Incident
			for _, incident := range incidents {
				if time.Since(incident.Time) < timebucket {
					newIncidents = append(newIncidents, incident)
				}
			}
			incidents = newIncidents

			// count violations per IP
			var buckets = make(map[string]int)
			for _, incident := range incidents {
				buckets[incident.Ip.String()] += 1
			}
			log.Println(buckets)
			for ip, violations := range buckets {
				if violations > max_incidents {
					blockChan <- blockstore.Block{Ip: net.ParseIP(ip), Time: time.Now(), External: false}
				}
			}
		}
	}
}
