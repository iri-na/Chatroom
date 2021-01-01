package chat

import (
	"Chat/backend/utils"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
)

type Chat struct {
	users    map[string]*User
	messages chan *Message
	join     chan *User
	leave    chan *User
}

var upgrader = websocket.Upgrader{
	ReadBufferSize: 512,
	WriteBufferSize: 512,
	CheckOrigin: func (r *http.Request) bool {
		log.Printf("%s %s%s %v\n", r.Method, r.Host, r.RequestURI, r.Proto)
		return r.Method == http.MethodGet
	},
}

func (c *Chat) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error on websocket connection: ", err.Error())
		return
	}

	keys := r.URL.Query()
	username := keys.Get("username")
	if strings.TrimSpace(username) == "" {
		username = fmt.Sprint("anom-%d", utils.GetRandomI64())
	}

	user := &User{
		Username: username,
		Conn: conn,
		Global: c,
	}

	c.join<- user

	user.Read()
}

func (c *Chat) Run() {
	for {
		select {
		case user := <-c.join:
			c.add(user)
		case message := <-c.messages:
			c.broadcast(message)
		case user := <-c.leave:
			c.disconnect(user)
		}
	}
}

func (c *Chat) add(user *User)  {
	if _, ok := c.users[user.Username]; !ok { //if this username is not in map
		c.users[user.Username] = user

		hello := fmt.Sprintf("%s has entered the chat. Total users in chat: %d\n", user.Username, len(c.users))
		c.broadcast(NewMessage(hello, "Server"))
	}
}

func (c *Chat) broadcast(message *Message) {
	log.Printf("Broadcast message: %v\n", message)
	//broadcast the message to all the users
	for _, user := range c.users {
		user.Write(message)
	}
}

func (c *Chat) disconnect(user *User) {
	if _, ok := c.users[user.Username]; ok { //if this username is not in map
		defer user.Conn.Close()
		delete(c.users, user.Username)

		goodbye := fmt.Sprintf("User %s has left the chat. Total users in chat: %d\n", user.Username, len(c.users))
		c.broadcast(NewMessage(goodbye, "Server"))
	}
}

func Start(port string) {
	log.Printf("Chat listening on http://localhost%s\n", port)

	//initialize Chat with nil values
	c := &Chat{
		users:    make(map[string]*User), //string is key, *User is value
		messages: make(chan *Message),
		join:     make(chan *User),
		leave:    make(chan *User),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Go Webchat!!"))
	})

	http.HandleFunc("/chat", c.Handler)

	go c.Run()

	log.Fatal(http.ListenAndServe(port, nil))
}
