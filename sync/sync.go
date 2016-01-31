package sync

import (
	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
	"github.com/timbuchwaldt/dblock/blocker"
	"log"
	"net"
	"strings"
	"time"
)

func Start(blockControlChan chan blocker.ControlMsg) {
	cfg := client.Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)
	go watchKey("dblock", blockControlChan, kapi)
	go watchKey("dblock6", blockControlChan, kapi)
}

func ipFromEtcdKey(key string) net.IP {
	split := strings.Split(key, "/")
	return net.ParseIP(split[2])
}

func watchKey(key string, blockControlChan chan blocker.ControlMsg, kapi client.KeysAPI) {
	watcher := kapi.Watcher(key, &client.WatcherOptions{Recursive: true})

	for {
		response, _ := watcher.Next(context.Background())

		switch response.Action {
		case "set":
			ip := ipFromEtcdKey(response.Node.Key)
			blockControlChan <- blocker.ControlMsg{Ip: ip, Block: true}
			log.Println("New key was set: " + ip.String())
		case "delete":
			ip := ipFromEtcdKey(response.Node.Key)
			blockControlChan <- blocker.ControlMsg{Ip: ip, Block: false}
			log.Println("key was deleted: " + response.Node.Key)
		case "expire":
			ip := ipFromEtcdKey(response.Node.Key)
			blockControlChan <- blocker.ControlMsg{Ip: ip, Block: false}
			log.Println("Key expired: " + response.Node.Key)
		}
	}
}
