package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Username string
	Conn     *websocket.Conn
}

type Packet struct {
	Type     string          `json:"type"`
	Username string          `json:"username,omitempty"`
	Sender   string          `json:"sender,omitempty"`
	Target   string          `json:"target,omitempty"`
	Payload  json.RawMessage `json:"payload,omitempty"`
}

var (
	clients = make(map[string]*Client)
	mutex   sync.RWMutex
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	defer conn.Close()

	var username string

	for {
		var packet Packet

		if err := conn.ReadJSON(&packet); err != nil {
			break
		}

		switch packet.Type {

		case "register":
			if packet.Username == "" {
				continue
			}

			username = packet.Username

			mutex.Lock()

			// Remove any previous connection using this username.
			if oldClient, exists := clients[username]; exists {
				_ = oldClient.Conn.Close()
			}

			clients[username] = &Client{
				Username: username,
				Conn:     conn,
			}

			mutex.Unlock()

			log.Println("[REGISTER]", username, "connected")

		case "offer", "answer", "candidate", "signal":
			if packet.Target == "" {
				continue
			}

			mutex.RLock()
			target, ok := clients[packet.Target]
			mutex.RUnlock()

			if !ok {
				log.Println(
					"[ROUTING] target not connected:",
					packet.Target,
				)
				continue
			}

			log.Println(
				"[FORWARD]",
				packet.Type,
				"from",
				packet.Sender,
				"to",
				packet.Target,
			)

			// The server only routes the packet.
			// It does not inspect or decrypt payload data.
			if err := target.Conn.WriteJSON(packet); err != nil {
				log.Println("[FORWARD ERROR]", err)
			}

		default:
			log.Println("unknown packet type:", packet.Type)
		}
	}

	if username != "" {
		mutex.Lock()

		// Only remove this connection if it is still
		// the active connection for the username.
		if client, exists := clients[username]; exists &&
			client.Conn == conn {
			delete(clients, username)
		}

		mutex.Unlock()

		log.Println(username, "disconnected")
	}
}

func main() {
	http.HandleFunc("/ws", handleConnection)

	// RunSite provides the PORT environment variable.
	// When running locally, default to port 8080.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("CRYPT signaling server running on :" + port)

	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}