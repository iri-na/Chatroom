package chat

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

type User struct {
	Username string
	Conn *websocket.Conn
	Global *Chat
}

func (u *User) Read() {
	for {
		// when there is no message, websocket waits at ReadMessage until there is a message to read
		if _, message, err := u.Conn.ReadMessage(); err != nil {
			log.Println("Error on read message: ", err.Error())
			break
		} else {
			u.Global.messages <- NewMessage(string(message), u.Username)
		}
	}
	// on break (when there is error reading message, like if u refresh)
	u.Global.leave <- u
}

func (u *User) Write(message *Message) {
	b, _ := json.Marshal(message)

	if err := u.Conn.WriteMessage(websocket.TextMessage, b); err != nil {
		log.Println(" Error on write message: ", err.Error())
	}
}