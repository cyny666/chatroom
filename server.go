package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// We'll need to define an UpGrader
// this will require a Read and Write buffer size
type User struct {
	ID   int
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

// 动态展示在线成员列表
func show_members(conn *websocket.Conn, userID string) {
	var usernames []string
	for _, user := range users {
		usernames = append(usernames, user.Name)

	}
	// 将切片转换为 JSON 字符串
	usernamesJSON, _ := json.Marshal(usernames)
	names := Message{
		Type:    "names",
		Content: string(usernamesJSON),
	}
	log.Println(names)
	// 发送消息到前端
	for _, u := range users {
		if err := u.Conn.WriteJSON(names); err != nil {
			log.Println(err)
			return
		}
	}

}
func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	// connection
	ws, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	// Generate a unique user ID (you can use UUID library or any other method)
	userID := len(users) + 1
	// Create a new User
	user := &User{
		ID:   userID,
		Conn: ws,
	}

	// Add the user to the users map

	users[strconv.Itoa(userID)] = user

	// say hello
	log.Printf("User %d connected\n", userID)
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
	go func(conn *websocket.Conn, userID int) {
		defer func() {
			// 离开时的工作

			var user_name = users[strconv.Itoa(userID)].Name
			delete(users, strconv.Itoa(userID))
			show_members(conn, strconv.Itoa(userID))
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
			// 向所有用户说一下该用户离开了
			for _, u := range users {
				if err := u.Conn.WriteMessage(1, jsonleaveMessage); err != nil {
					log.Println(err)
					return
				}

				conn.Close()
			}
		}()
		// 读取用户的名称(此时的数据为JSon格式）
		_, nickname, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		var nickname_message Message
		// 使用 json.Unmarshal 解析 JSON 数据
		if err := json.Unmarshal(nickname, &nickname_message); err != nil {
			log.Println(err)
			return
		}
		// 设置用户的名称
		users[strconv.Itoa(userID)].Name = nickname_message.Content
		show_members(conn, strconv.Itoa(userID))
		// 将该用户的名称广播给其他
		enterMessage := Message{
			Type:    "text",
			Content: fmt.Sprintf("User %s joined the chat\n", users[strconv.Itoa(userID)].Name),
		}
		// 将 Message 结构体编码为 JSON
		jsonenterMessage, err := json.Marshal(enterMessage)
		if err != nil {
			log.Println(err)
			return
		}

		for _, u := range users {
			if err := u.Conn.WriteMessage(1, jsonenterMessage); err != nil {
				log.Println(err)
				return
			}
		}

		for {

			// read in a message
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			//判断收到信息的种类
			var message Message
			if err := json.Unmarshal(p, &message); err != nil {
				log.Println(err)
				return
			}
			if message.Type == "text" { // 将message广播出去
				log.Printf("[%s] %s\n", nickname_message.Content, string(message.Content))
				MessageTypestr := Message{
					Type:    "text",
					Content: fmt.Sprintf("%s%s", nickname_message.Content, ":") + string(message.Content),
				}
				// 将 Message 结构体编码为 JSON
				jsonMessageTypestr, _ := json.Marshal(MessageTypestr)
				// 将消息发送给其他所有用户
				for _, u := range users {
					if err := u.Conn.WriteMessage(1, jsonMessageTypestr); err != nil {
						log.Println(err)
						return
					}
					log.Println(jsonMessageTypestr)

				}

			} else {
				private_message := Message{
					Type:    "text",
					Content: fmt.Sprintf("%s%s", "(私聊)"+nickname_message.Content, ":") + string(message.Content),
				}
				json_private, err := json.Marshal(private_message)
				if err != nil {
					log.Println("JSON编组错误:", err)
					// 根据情况处理错误，例如返回或记录并继续
					continue
				}

				for _, u := range users {
					if u.Name == message.Type || u.ID == userID {
						// 使用WriteMessage发送字节数组而非WriteJSON
						err := u.Conn.WriteMessage(websocket.TextMessage, json_private)
						if err != nil {
							log.Println("发送消息错误:", err)
						}
						log.Println(string(json_private))
					}
				}
			}

			log.Println("finished")

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
