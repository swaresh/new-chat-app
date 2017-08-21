package main

import (
	"log"
	"net/http"
	"bytes"
	"strings"
	"os/exec"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel
var usernames = make(map[*websocket.Conn]string)
// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Define our message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	// Create a simple file server
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func dosomething(msg Message, client *websocket.Conn) {
	line := msg.Message
	to_send := &Message{
		Email: "email",
		Username: "bot",
		Message: "",
	}
	cmd := exec.Command("")
	command := ""
	switch line {
			case "build agent" :
				cmd = exec.Command("go", "run", "agent.go")
			case "build asset" :
				cmd = exec.Command("go", "run", "asset.go")
			case "build performance" :
				cmd = exec.Command("go", "run", "performance.go")
				default: command = "invalid-input"
			}
				if command=="invalid-input"{
					to_send.Message = "I am not able to process your request. Please choose from\n1.build agent 2.build asset 3.build performance\n"
				} else {
				cmd.Stdin = strings.NewReader("")
				var out bytes.Buffer
				cmd.Stdout = &out
				err := cmd.Run()
				if err != nil {
					log.Fatal(err)
				}
				parts := strings.Split(out.String(), "\n")
				//client.outgoing <- "\n"
				for _, part := range parts{
					to_send.Message = to_send.Message + part + "\n"
					//to_send.Message = "\n"
				}
				// client.outgoing <- "\n"
			}
			err := client.WriteJSON(to_send)
			//dosomething(to_send, client)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}

}



func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
		usernames[ws] = msg.Username
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			if(usernames[client]==msg.Username){
			err := client.WriteJSON(msg)
			dosomething(msg, client)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		}
	}
}
