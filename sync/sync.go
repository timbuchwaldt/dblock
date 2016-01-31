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

	log.Println("[sync]\tEnsuring dblock + dblock6 directories exist")
	kapi.Set(context.Background(), "dblock", "", &client.SetOptions{Dir: true})
	kapi.Set(context.Background(), "dblock6", "", &client.SetOptions{Dir: true})

	log.Println("[sync]\tReading dblock entries")
	result, err2 := kapi.Get(context.Background(), "dblock", &client.GetOptions{Recursive: true})
	if err2 != nil {
		log.Fatal(err2)
	}

	for _, node := range result.Node.Nodes {
		handleKey(*node, blockControlChan, true)
	}

	log.Println("[sync]\tReading dblock6 entries")
	result6, err3 := kapi.Get(context.Background(), "dblock6", &client.GetOptions{Recursive: true})
	if err3 != nil {
		log.Fatal(err3)
	}

	for _, node := range result6.Node.Nodes {
		handleKey(*node, blockControlChan, true)
	}

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
			handleKey(*response.Node, blockControlChan, true)
			log.Println("[sync]\tetcd: add: " + response.Node.Key)
		case "delete":
			handleKey(*response.Node, blockControlChan, false)
			log.Println("[sync]\tetcd: delete: " + response.Node.Key)
		case "expire":
			handleKey(*response.Node, blockControlChan, false)
			log.Println("[sync]\tetcd: expired: " + response.Node.Key)
		}
	}
}

func handleKey(node client.Node, blockControlChan chan blocker.ControlMsg, block bool) {
	ip := ipFromEtcdKey(node.Key)
	blockControlChan <- blocker.ControlMsg{Ip: ip, Block: block}
}
