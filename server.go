package main

import (
	"encoding/json"
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
	Name string
}
type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
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
	nameMessage := Message{
		Type:    "text",
		Content: "请输入你的昵称",
	}
	// 将 Message 结构体编码为 JSON
	jsonNameMessage, err := json.Marshal(nameMessage)
	if err != nil {
		log.Println(err)
		return
	}

	// 使用 WriteMessage 发送 JSON 编码的消息
	err = ws.WriteMessage(1, jsonNameMessage)
	if err != nil {
		log.Println(err)
		return
	}

	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	go func(conn *websocket.Conn, userID string) {
		defer func() {
			// Remove the user from the users map when the connection is closed
			usersLock.Lock()
			var user_name = users[userID].Name
			delete(users, userID)
			usersLock.Unlock()

			leaveMessage := Message{
				Type:    "text",
				Content: fmt.Sprintf("User %s disconnected\n", user_name),
			}
			// 将 Message 结构体编码为 JSON
			jsonleaveMessage, err := json.Marshal(leaveMessage)
			if err != nil {
				log.Println(err)
				return
			}
			for _, u := range users {
				if err := u.Conn.WriteMessage(1, jsonleaveMessage); err != nil {
					log.Println(err)
					return
				}

				conn.Close()
			}
		}()
		// Read the user's nickname
		_, nickname, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// Set the user's nickname
		usersLock.Lock()
		users[userID].Name = string(nickname)
		usersLock.Unlock()
		// Broadcast the user's entrance message
		enterMessage := Message{
			Type:    "text",
			Content: fmt.Sprintf("User %s joined the chat\n", users[userID].Name),
		}
		// 将 Message 结构体编码为 JSON
		jsonenterMessage, err := json.Marshal(enterMessage)
		if err != nil {
			log.Println(err)
			return
		}
		usersLock.Lock()
		for _, u := range users {
			if err := u.Conn.WriteMessage(1, jsonenterMessage); err != nil {
				log.Println(err)
				return
			}
		}
		//将numberlist更新一下
		for _, u := range users {
			if err := u.Conn.WriteMessage(websocket.TextMessage, []byte("number:"+userID)); err != nil {
				log.Println(err)
				return
			}
		}
		usersLock.Unlock()
		for {

			// read in a message
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			// print out that message for clarity
			log.Printf("[%s] %s\n", nickname, string(p))
			MessageTypestr := Message{
				Type:    "text",
				Content: fmt.Sprintf("%s%s", nickname, ":") + string(p),
			}
			// 将 Message 结构体编码为 JSON
			jsonMessageTypestr, err := json.Marshal(MessageTypestr)
			// Broadcast the message to all other users
			usersLock.Lock()
			for _, u := range users {
				if err := u.Conn.WriteMessage(1, jsonMessageTypestr); err != nil {
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
