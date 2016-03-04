package incidentstore

import (
	"github.com/timbuchwaldt/dblock/blocker"
	"github.com/timbuchwaldt/dblock/config"
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

func IncidentStore(incidentChan chan Incident, syncChannel chan blocker.ControlMsg, config config.Config) {
	log.Println(config)
	ticker := time.NewTicker(time.Second * 5)
	var incidents []Incident
	for {
		select {
		case msg := <-incidentChan:
			// counter: incidentstore.incident_received
			incidents = append(incidents, msg)
		case <-ticker.C:
			// counter: incidentstore.cleanup
			// timer: incidentstore.cleanup
			start := time.Now()
			var newIncidents []Incident
			for _, incident := range incidents {
				if time.Since(incident.Time) < time.Duration(config.IncidentTime.Nanoseconds()) {
					newIncidents = append(newIncidents, incident)
				}
			}
			incidents = newIncidents

			// count violations per IP
			var buckets = make(map[string]int)
			for _, incident := range incidents {
				buckets[incident.Ip.String()] += 1
			}
			for ip, violations := range buckets {
				if violations > config.MaxIncidents {
					// counter: incidentstore.blocks
					syncChannel <- blocker.ControlMsg{Ip: net.ParseIP(ip), Block: true}
				}
			}
			elapsed := time.Since(start)
			log.Printf("cleanup took %s", elapsed)
		}
	}
}
