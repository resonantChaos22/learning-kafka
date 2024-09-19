package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"topic-two/items"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (e *Executer) Stream() {
	defer e.wg.Done()

	color.Green("Starting Stream...")

	store, err := items.NewPostgresStore()
	if err != nil {
		log.Fatalf("Error in creating Postgres Store: %v", err)
	}
	e.store = store

	// Channel to send values
	itemChan := make(chan items.Item)
	wgStream := new(sync.WaitGroup)

	// WebSocket handler
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		// Extract itemID from query parameters
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

		// Start listening for changes for the provided itemID
		wgStream.Add(1)
		go e.cluster.ListenForItemChanges(itemID, itemChan, wgStream, e.ctx)

		for {
			select {
			case <-e.ctx.Done():
				wgStream.Wait()
				close(itemChan)
				return
			case item := <-itemChan:
				if item.ID == itemID {
					log.Println("Sending value:", item.Value)

					// Send value to WebSocket clients
					err := conn.WriteJSON(item.Value)
					if err != nil {
						log.Println("Failed to send message:", err)
						return
					}
				}
			}
		}
	})

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade HTTP to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade failed:", err)
			return
		}
		defer conn.Close()

		// Example: Read and respond to WebSocket messages
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading WebSocket message:", err)
				break
			}
			log.Printf("Received: %s", message)

			// Send a response back to the client
			err = conn.WriteMessage(websocket.TextMessage, []byte("Hello, Client!"))
			if err != nil {
				log.Println("Error writing WebSocket message:", err)
				break
			}
		}
	})

	// Start WebSocket server
	server := &http.Server{Addr: ":8001"}

	go func() {
		color.Green("WebSocket server started on :8001")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("WebSocket server failed: %v", err)
		}
	}()

	// Graceful shutdown when context is canceled
	<-e.ctx.Done()
	log.Println("Shutting down WebSocket server...")
	server.Shutdown(e.ctx)
	wgStream.Wait()
}
