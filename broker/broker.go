package broker

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ViktorEdman/gssc-go/data"
	"github.com/ViktorEdman/gssc-go/helpers"
	"github.com/ViktorEdman/gssc-go/templates"
)

const (
	timeout           time.Duration = time.Second * 1
	heartBeatInterval time.Duration = 30 * time.Second
)

type SSEEvent struct {
	Event string
	Data  string
}

func (e SSEEvent) String() string {
	return fmt.Sprintf("event: %s\ndata:%s\n\n", e.Event, e.Data)
}

type Broker struct {
	Events         chan SSEEvent
	newClients     chan chan SSEEvent
	closingClients chan chan SSEEvent
	clients        map[chan SSEEvent]bool
	db             *data.Queries
}

func NewBroker(db *data.Queries) (broker *Broker) {
	broker = &Broker{
		Events:         make(chan SSEEvent),
		newClients:     make(chan chan SSEEvent),
		closingClients: make(chan chan SSEEvent),
		clients:        make(map[chan SSEEvent]bool),
		db:             db,
	}
	go broker.run()
	return broker
}

func (broker *Broker) AddClient(clientChan chan SSEEvent) {
	broker.newClients <- clientChan
}

func (broker *Broker) RemoveClient(clientChan chan SSEEvent) {
	broker.closingClients <- clientChan
}

func (broker *Broker) BroadcastEvent(event SSEEvent) {
	broker.Events <- event
}

func (broker *Broker) run() {
	for {
		select {
		case s := <-broker.newClients:
			broker.clients[s] = true
		case s := <-broker.closingClients:
			delete(broker.clients, s)
		case event := <-broker.Events:
			if len(broker.clients) > 0 {
				log.Printf("Notifying %d clients.\nMessage: %s", len(broker.clients), event.Event)
			}

			for clientChannel := range broker.clients {
				go func(clientChannel chan SSEEvent) {
					select {
					case clientChannel <- event:
					case <-time.After(timeout):
						log.Println("Skipping client")
					}
				}(clientChannel)
			}
		}
	}
}

func (broker Broker) SSEHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	msgChan := make(chan SSEEvent)
	broker.AddClient(msgChan)
	defer func() {
		broker.RemoveClient(msgChan)
		close(msgChan)
	}()
	ctx := r.Context()
	heartBeatTimer := time.NewTicker(heartBeatInterval)
	now := time.Now()
	timeString := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second())
	fmt.Fprint(w, SSEEvent{
		Event: "Connected",
		Data:  timeString,
	})
	log.Println("Client connected")
	flusher.Flush()
	for {
		select {
		case <-ctx.Done():
			heartBeatTimer.Stop()
			return
		case message := <-msgChan:
			if strings.Contains(message.Event, "server-") {
				if message.Data != "deleted" {
					event := strings.Split(message.Event, "-")
					serverId, err := strconv.ParseInt(event[1], 10, 64)
					if err != nil {
						log.Println(err)
						continue
					}
					serverData, err := helpers.GetLatestStatusWithPlayer(serverId, broker.db)
					if err != nil {
						log.Println(err)
						continue
					}
					template := templates.ServerTemplate(*serverData)
					buf := bytes.Buffer{}
					template.Render(r.Context(), &buf)
					message.Data = buf.String()
				} else {
					message.Data = ""
				}
			}
			if strings.Contains(message.Event, "newserver") {
				serverId, err := strconv.ParseInt(message.Data, 10, 64)
				if err != nil {
					log.Println(err)
					continue
				}
				serverData, err := helpers.GetLatestStatusWithPlayer(serverId, broker.db)
				if err != nil {
					log.Println(err)
					continue
				}
				template := templates.ServerTemplate(*serverData)
				buf := bytes.Buffer{}
				template.Render(r.Context(), &buf)
				message.Data = buf.String()
			}
			fmt.Fprint(w, message.String())
			flusher.Flush()
		case time := <-heartBeatTimer.C:
			fmt.Fprint(w, SSEEvent{
				Event: "heartbeat",
				Data: fmt.Sprintf(
					"%d-%02d-%02d %02d:%02d:%02d",
					time.Year(),
					time.Month(),
					time.Day(),
					time.Hour(),
					time.Minute(),
					time.Second(),
				),
			})
			flusher.Flush()
		}
	}
}
