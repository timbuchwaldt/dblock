package dblockmain

import (
	"flag"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/timbuchwaldt/dblock/blocker"
	"github.com/timbuchwaldt/dblock/incidentstore"
	"github.com/timbuchwaldt/dblock/sync"

	"log"
	"net"
	"net/http"
	"regexp"
	"time"
)

var (
	addr          = flag.String("listen-address", ":8080", "The address to listen for prometheus requests.")
	timebucket    = flag.Duration("incident-bucket", 5*time.Minute, "The number of seconds of incidents we compare.")
	max_incidents = flag.Int("max-incidents", 5, "The number incidents allowed per incident-bucket.")
	config_file   = flag.String("config", "/etc/dblock.toml", "The TOML config file to read.")
)

func Main() {
	flag.Parse()
	config := ParseConfig(*config_file)

	http.Handle("/metrics", prometheus.Handler())

	incidentChan := make(chan incidentstore.Incident, 100)
	blockControlChan := make(chan blocker.ControlMsg, 100)
	syncChannel := make(chan blocker.ControlMsg, 100)

	/*
		Startup: Start blocker, block store, incident store
	*/
	go blocker.Blocker(blockControlChan)
	go incidentstore.IncidentStore(incidentChan, syncChannel, *timebucket, *max_incidents)
	go sync.Start(blockControlChan, syncChannel)
	/*
		Startup: start log file follower based on toml config
	*/

	for name, fileConfig := range config.Files {
		log.Println("[main]\tSetting up follower for " + name)
		go follow_and_analyze(fileConfig.Filename, fileConfig.Regexes, incidentChan)
	}

	http.ListenAndServe(*addr, nil)

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

func follow_and_analyze(filename string, regexes []string, c chan incidentstore.Incident) {
	t, err := tail.TailFile(filename,
		tail.Config{
			Follow:   true,                                  // actually follow the logs
			ReOpen:   true,                                  // allow logs to be rotated
			Location: &tail.SeekInfo{Offset: 0, Whence: 2}}) // seek to end of file

	if err != nil {
		log.Fatal(err)
	}
	var compiledRegexes []regexp.Regexp
	for _, regex := range regexes {
		compiledRegex := regexp.MustCompile(regex)
		compiledRegexes = append(compiledRegexes, *compiledRegex)
	}

	for line := range t.Lines {
		// match each line against all regexes
		for _, regex := range compiledRegexes {
			result := regex.FindStringSubmatch(line.Text)
			if result != nil {
				ip := net.ParseIP(result[1])
				if ip != nil {
					c <- incidentstore.Incident{Filename: filename, Ip: ip, Time: time.Now(), Line: line.Text}

					// break here, this line matched on regex
					break
				}
			}
		}
		// match against regexes here, if one matches, create "incident" struct
	}

}
