package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"topic-two/items"
	"topic-two/kafka/handlers"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type StreamMessage struct {
	Value     float64 `json:"value"`
	Timestamp int     `json:"time"`
	ID        int     `json:"id"`
}

func (e *Executer) Stream() {
	defer e.wg.Done()

	color.Green("Starting Stream...")

	store, err := items.NewPostgresStore()
	if err != nil {
		log.Fatalf("Error in creating Postgres Store: %v", err)
	}
	e.store = store

	itemChan := make(chan handlers.DebeziumUpdateMessage)
	wgStream := new(sync.WaitGroup)

	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Connected!")
		itemIDStr := r.URL.Query().Get("itemID")
		if itemIDStr == "" {
			http.Error(w, "itemID is required", http.StatusBadRequest)
			return
		}

		itemID, err := strconv.Atoi(itemIDStr)
		if err != nil {
			http.Error(w, "Invalid itemID", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade to WebSocket:", err)
			return
		}
		defer conn.Close()

		wgStream.Add(1)
		go e.cluster.ListenForItemChanges(itemID, itemChan, wgStream, e.ctx)

		for {
			select {
			case <-e.ctx.Done():
				wgStream.Wait()
				close(itemChan)
				return
			case msg := <-itemChan:
				if msg.Item.ID == itemID {
					log.Println("Sending value:", msg.Item.Value)

					err := conn.WriteJSON(NewStreamMessage(msg.Item.ID, msg.TimeStamp, msg.Item.Value))
					if err != nil {
						log.Println("Failed to send message:", err)
						wgStream.Wait()
						close(itemChan)
						return
					}
				}
			}
		}
	})

	server := &http.Server{Addr: ":8001"}

	go func() {
		color.Green("WebSocket server started on :8001")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("WebSocket server failed: %v", err)
		}
	}()

	<-e.ctx.Done()
	log.Println("Shutting down WebSocket server...")
	server.Shutdown(e.ctx)
	wgStream.Wait()
}

func NewStreamMessage(id, timestamp int, value float64) StreamMessage {
	return StreamMessage{
		ID:        id,
		Value:     value,
		Timestamp: timestamp,
	}
}
