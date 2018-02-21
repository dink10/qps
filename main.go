package main

import (
	"encoding/json"
	"flag"
	"github.com/dink10/go-wsqueue"
	"github.com/gorilla/mux"
	"github.com/paulbellamy/ratecounter"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	PROTOCOL = "ws"
	PATH     = "/"
)

var (
	config  ConfigType
	fServer = flag.Bool("server", false, "Run server")
	fClient = flag.Bool("client", false, "Run client")
)

type Message struct {
	Id int
}

var CampaignsCounters = make(map[int]*ratecounter.RateCounter)

func main() {

	config, _ = GetConfig()
	flag.Parse()
	reconfigure()

	serverId := GetServerId(config.Server.Host, config.Server.Topic)

	if *fServer {
		server(config.Server.Host, config.Server.Topic)
	}
	if *fClient {
		client(serverId, config.Nodes)
	}

	select {}
}

// Reconfigure application by SIGHUP
// Example kill -SIGHUP 1414
func reconfigure() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP)
	go func(c chan os.Signal) {
		for {
			switch <-c {
			case syscall.SIGHUP:
				var err error
				config, err = ReloadConfig()
				log.Println("SIGHUP received. Going to reload config ...")
				if err != nil {
					log.Printf("error while reloading config: %s", err)
					continue
				}
				log.Println("Reloading config: successful")
			}
		}
	}(c)
}

// Start server
func server(server string, topic string) {

	r := mux.NewRouter()
	s := wsqueue.NewServer(r, "")
	q := s.CreateTopic(topic)
	r.HandleFunc("/", getHandleFunc(q))

	q.OpenedConnectionHandler = func(c *wsqueue.Conn) {
		q.Publish("Welcome " + c.ID)
	}

	q.ClosedConnectionHandler = func(c *wsqueue.Conn) {
		log.Println("Bye bye " + c.ID)
	}
	http.Handle(PATH, r)
	go http.ListenAndServe(server, r)

}

// Connect a client
func client(serverId int, nodes Nodes) {
	for _, node := range nodes {
		nodeId := GetServerId(node.Host, node.Topic)
		log.Println(int(serverId) != int(nodeId))
		if int(serverId) != int(nodeId) {
			go runClient(node, serverId)
		}
	}
}

// Http server handler
func getHandleFunc(q *wsqueue.Topic) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if campaignId, err := validateParams(r); !err {
			go func() {
				q.Publish(Message{Id: campaignId})
			}()
		}
	}
}

// Validation request params
func validateParams(r *http.Request) (int, bool) {
	key := r.URL.Query().Get("id")

	if len(key) == 0 {
		log.Println("Url Param ID is missing")
		return 0, false
	}

	id, err := strconv.Atoi(key)

	if err != nil {
		log.Printf("Url Param ID error: %s", err)
		return 0, false
	}

	return id, true
}

// Run a client
func runClient(node Node, serverId int) {
	c := &wsqueue.Client{
		Protocol: PROTOCOL,
		Host:     node.Host,
		Route:    PATH,
	}
	cMessage, cError, err := c.Subscribe(node.Topic)
	if err != nil {
		panic(err)
	}
	for {
		select {
		case m := <-cMessage:
			var message Message
			err = json.Unmarshal([]byte(m.Body), &message)
			if err == nil {
				if _, ok := CampaignsCounters[message.Id]; !ok {
					CampaignsCounters[message.Id] = NewCounter()
				}
				Increment(CampaignsCounters[message.Id])
				log.Println(Count(CampaignsCounters[message.Id]))
			}
		case e := <-cError:
			log.Println("\n\n********* Client " + string(serverId) + "  *********" + e.Error() + "\n******************")
		}
	}
}
