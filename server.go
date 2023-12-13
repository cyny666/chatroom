package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

// We'll need to define an UpGrader
// this will require a Read and Write buffer size
type User struct {
	ID   string
	Conn *websocket.Conn
}

const MessageTypeContext = 1000

var (
	upGrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	users     = make(map[string]*User)
	usersLock sync.Mutex
)

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	// connection
	ws, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	// Generate a unique user ID (you can use UUID library or any other method)
	userID := fmt.Sprintf("user_%d", len(users)+1)
	// Create a new User
	user := &User{
		ID:   userID,
		Conn: ws,
	}

	// Add the user to the users map
	usersLock.Lock()
	users[userID] = user
	usersLock.Unlock()

	// say hello
	log.Printf("User %s connected\n", userID)
	err = ws.WriteMessage(1, []byte(fmt.Sprintf("Hi, %s!", userID)))
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	go func(conn *websocket.Conn, userID string) {
		defer func() {
			// Remove the user from the users map when the connection is closed
			usersLock.Lock()
			delete(users, userID)
			usersLock.Unlock()
			leaveMessage := fmt.Sprintf("User %s disconnected\n", userID)
			//这里定义1000为用户数量表
			for _, u := range users {
				if err := u.Conn.WriteMessage(1, []byte(leaveMessage)); err != nil {
					log.Println(err)
					return
				}

				conn.Close()
			}
		}()

		for {
			// read in a message
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			// print out that message for clarity
			log.Printf("[%s] %s\n", userID, string(p))
			messageTypeStr := fmt.Sprintf("%s%s", userID, ":")
			// Broadcast the message to all other users
			usersLock.Lock()
			for _, u := range users {
				if err := u.Conn.WriteMessage(messageType, []byte(messageTypeStr+string(p))); err != nil {
					log.Println(err)
					return
				}

			}
			usersLock.Unlock()
		}
	}(ws, userID)

}

func setupRoutes() {
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("starting websocket chatroom!")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
