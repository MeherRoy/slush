package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{} // use default websocket options

// Data to stream to clients
type ClientData struct {
	Id    int    // node (player) id
	Color string // node color
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print(r.Host, " connected (ws)")

	// upgrade the http connection to a websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	// starts the simulation once the connection is established (few lines back)
	go sl.startSimulation()
	defer c.Close()

	// send some messages to the client
	for {
		data := <- colorChange
		buf, _ := json.Marshal(data)

		err = c.WriteMessage(websocket.TextMessage, buf)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}

	// Echo client messages	(write what we read)
	/*for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}*/
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Print(r.Host, " connected")
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "index.html")
}
