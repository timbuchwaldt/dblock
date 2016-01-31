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

func Start(blockControlChan chan blocker.ControlMsg, syncChannel chan blocker.ControlMsg) {
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
	go sync(kapi, syncChannel)
}

func sync(kapi client.KeysAPI, syncChannel chan blocker.ControlMsg) {
	for {
		msg := <-syncChannel
		var folder string
		if msg.Ip.To4() != nil {
			folder = "dblock6/"
		} else {
			folder = "dblock/"
		}
		_, err := kapi.Set(context.Background(), folder+msg.Ip.String(), "0", &client.SetOptions{TTL: 5 * time.Second})
		if err != nil {
			log.Fatal(err)
		} else {
			// print common key info
			log.Println("[sync]\tAdded block: " + msg.Ip.String())
		}
	}

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
			log.Println("[sync]\tetcd: add" + ip.String())
		case "delete":
			ip := ipFromEtcdKey(response.Node.Key)
			blockControlChan <- blocker.ControlMsg{Ip: ip, Block: false}
			log.Println("[sync]\tetcd: delete: " + response.Node.Key)
		case "expire":
			ip := ipFromEtcdKey(response.Node.Key)
			blockControlChan <- blocker.ControlMsg{Ip: ip, Block: false}
			log.Println("[sync]\tetcd: expired: " + response.Node.Key)
		}
	}
}
