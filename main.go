package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type Message struct {
	Content string `json:"content"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

func main() {
	// setup file server to serve html files
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	//configure websocket route
	http.HandleFunc("/ws", handleConnections)

	//start listening for incoming chat messages
	go handleMessages()
	go readUserInput()

	fmt.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()

	// register new client
	clients[ws] = true

	for {
		//Read in a new message as JSON and map it to a Message object
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading JSON", err)
			delete(clients, ws)
			break
		}

		fmt.Printf("\nReceived Message: %s\n", msg.Content)
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println("Error writing JSON:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func readUserInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		//prompt user for input
		fmt.Print("Enter message to send to clients: ")
		text, _ := reader.ReadString('\n')

		//send user input to broadcast channel
		msg := Message{Content: text}
		broadcast <- msg
	}
}
