package endpoints

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	failedPongCount int
	upgrader        = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients         sync.Map
	pongMutex       sync.Mutex
	clientPool      sync.Pool
)

const (
	updateInterval = 5 * time.Second
	pongWait       = 5 * time.Second
	maxRetries     = 2
	retryInterval  = 5 * time.Second
)

type RetryableOperation func() (interface{}, error)

type Client struct {
	Conn        *websocket.Conn
	UserID      string
	IsConnected bool
	mu          sync.Mutex
}

func init() {
	clientPool = sync.Pool{
		New: func() interface{} {
			return &Client{}
		},
	}
}

// retrying any func if needed
// func retryOperationWithTimeout(operation RetryableOperation, maxRetries int, retryInterval time.Duration) (interface{}, error) {
// 	var result interface{}
// 	var err error

// 	for i := 0; i < maxRetries; i++ {
// 		result, err = operation()
// 		if err == nil {
// 			return result, nil
// 		}

// 		log.Printf("Error: %v. Retrying in %s...", err, retryInterval)
// 		time.Sleep(retryInterval)
// 	}

// 	return nil, fmt.Errorf("failed after %d retries: %v", maxRetries, err)
// }

func getClient() *Client {
	return clientPool.Get().(*Client)
}

func putClient(client *Client) {
	clientPool.Put(client)
}

func isClientConnected(client *Client) bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.IsConnected
}

func handleClientDisconnection(client *Client, userID string, conn *websocket.Conn) {
	client.mu.Lock()
	defer client.mu.Unlock()

	if !client.IsConnected {
		log.Printf("Client is already disconnected: %s, UserID: %s\n", client.Conn.RemoteAddr(), userID)
		return
	}

	client.IsConnected = false
	clients.Delete(client)
	if err := conn.Close(); err != nil {
		log.Printf("Error closing WebSocket connection for client %s: %v", userID, err)
	}

	log.Printf("Client disconnected: %s, UserID: %s\n", client.Conn.RemoteAddr(), userID)
}

func (h *APIServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading the HTTP: %v\n", err)
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}

	userIDKey := "patluPrasadDevkota" // this should be passed from context after successfull authentication of token.
	userID := r.Context().Value(userIDKey).(string)
	if userID == "" {
		http.Error(w, "No user ID found in the context", http.StatusInternalServerError)
		return
	}

	client := getClient()

	client.Conn = conn
	client.UserID = userID
	client.IsConnected = true

	defer func() {
		putClient(client)
		handleClientDisconnection(client, userID, conn)
	}()

	log.Printf("Client connected: %s, UserID: %s\n", conn.RemoteAddr(), userID)
	log.Printf("Connected Client status: %+v\n", client)

	// DEconding the json data from client.

	// _, message, err := conn.ReadMessage()
	// if err != nil {
	// 	log.Printf("Error reading message from client: %v\n", err)
	// 	return
	// }

	// if err := json.Unmarshal(message, &type of data); err != nil {
	// fmt.Println("marshalling error")
	// log.Printf("Error decoding JSON message: %v\n", err)
	// return
	// }

	// separate goroutine for recieving any new message from client.
	// this goroutine will keep listening for any new message from client.
	go func() {
		defer handleClientDisconnection(client, userID, conn)
		for {
			// unmarshalling the json message from client
			// _, message, err := conn.ReadMessage()
			// if err != nil {
			// 	log.Printf("Error reading message from client: %v\n", err)
			// 	return
			// }
			// if err := json.Unmarshal(message, &type of data); err != nil {
			// 	fmt.Println("marshalling error")
			// 	log.Printf("Error decoding JSON message: %v\n", err)
			// 	return
			// }

			// for using the retry func,
			// retryableUpdateLocation := func() (interface{}, error) {
			//	return nil, a func
			// }
			return
		}
	}()

	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	pingTicker := time.NewTicker(pongWait)
	defer pingTicker.Stop()

	pongTicker := time.NewTicker(pongWait)
	defer pongTicker.Stop()

	// a separate goroutine for pingpong.
	go func() {
		defer handleClientDisconnection(client, userID, conn)
		for range pongTicker.C {
			if isClientConnected(client) {
				pongMutex.Lock()
				failedPongCount = 0
				log.Println("Received pong from client")
				pongMutex.Unlock()
			}
		}
	}()

	// sending data from server in ticker interval.
	for {
		select {
		case <-ticker.C:
			if !isClientConnected(client) {
				return
			}

			// sending the data

			// if !isEmptyLocationData(locationData) && isClientConnected(client) {
			// 	if err := h.sendNearbyUsers(client, locationData); err != nil {
			// 		log.Printf("Error sending nearby users to client %s: %v", userID, err)
			// 		handleClientDisconnection(client, userID, conn)
			// 		return
			// 	}
			// }

			// if server dont recieve pong from client more than 3 times it will disconnect the client.
			if failedPongCount >= maxRetries {
				if !isClientConnected(client) {
					return
				}
				handleClientDisconnection(client, userID, conn)
				return
			}

		case <-pingTicker.C:
			if !isClientConnected(client) {
				return
			}
			err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(pongWait))
			if err != nil {
				return
			}
			log.Printf("Sent ping to client")
		}
	}
}
