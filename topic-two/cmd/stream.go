package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
	"topic-two/kafka"

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

	wgStream := new(sync.WaitGroup)
	numConn := 0

	broadcast := kafka.NewItemsBroadcast()

	wgStream.Add(1)
	go e.cluster.ListenForAllItemChanges(broadcast, wgStream, e.ctx, e.errChan)

	http.HandleFunc("/stream", e.streamHandler(&numConn, wgStream, broadcast))

	server := &http.Server{Addr: ":8001"}

	go func() {
		color.Green("WebSocket server started on :8001")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.errChan <- fmt.Errorf("websocket server failed: %v", err)
		}
	}()

	<-e.ctx.Done()
	color.Red("Shutting down WebSocket server...")
	server.Shutdown(e.ctx)
	wgStream.Wait()
}

func (e *Executer) streamHandler(numConn *int, wgStream *sync.WaitGroup, broadcast *kafka.ItemsBroadcast) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := *numConn + 1
		*numConn++
		color.Green("User#%d Connected!", id)
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

		itemChan := broadcast.Register(itemID)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade to WebSocket:", err)
			return
		}
		defer conn.Close()

		wgStream.Add(1)
		retries := 1

		closeConn := func() {
			broadcast.Unregister(itemID, itemChan)
			color.Red("Closing connection for User#%d", id)
			wgStream.Done()
		}

		for {
			select {
			case <-e.ctx.Done():
				closeConn()
				return
			case msg, ok := <-itemChan:
				if !ok {
					log.Printf("Channel closed, exiting stream loop for User#%d\n", id)
					wgStream.Done()
					return
				}
				if msg.Item.ID == itemID || itemID == 0 {

					err := conn.WriteJSON(NewStreamMessage(msg.Item.ID, msg.TimeStamp, msg.Item.Value))
					if err != nil {
						color.Green("Retry#%d to send message to User#%d...\n", retries, id)
						retries++
						time.Sleep(1 * time.Second)
						if retries <= 5 {
							continue
						}
						color.Red("Failed to send message, closing connection for User#%d: %v", id, err)
						closeConn()
						return
					} else {
						if retries > 1 {
							color.Green("Regained Connection for User#%d!", id)
						}
						retries = 1
					}

					color.HiBlue("Sending value for User#%d and %s :%f\n", id, msg.Item.Name, msg.Item.Value)
				}
			}
		}
	}
}

func NewStreamMessage(id, timestamp int, value float64) StreamMessage {
	return StreamMessage{
		ID:        id,
		Value:     value,
		Timestamp: timestamp,
	}
}
