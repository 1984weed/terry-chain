package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// var clients = make(map[*websocket.Conn]bool)
// var broadcast = make(chan []byte)
// var upgrader = websocket.Upgrader{
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true
// 	},
// }

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "blockchain.html")
	})
	http.HandleFunc("/blocks", blocksHandler)
	http.HandleFunc("/mineBlock", mineBlock)
	// http.HandleFunc("/peers", func(w http.ResponseWriter, r *http.Request) {
	// 	response := make([]string{}, len(clients))
	// 	i := 0
	// 	for ws, value := range clients {
	// 		if value == true {
	// 			response[i] = ws.RemoteAddr().String()
	// 			i += 1
	// 		}
	// 	}
	// 	json.NewEncoder(w).Encode(response)
	// })
	hub := newHub()

	go hub.run()

	http.HandleFunc("/addPeer", blocksHandler)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	http.ListenAndServe(":8080", nil)
}

func blocksHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(GetBlockChain())
}

func mineBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		results := string(body)
		json.NewEncoder(w).Encode(GenerateNextBlock(results))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

}

// func handleConnections(w http.ResponseWriter, r *http.Request) {
// 	serveWs(hub, w, r)
// 	// ws, err := upgrader.Upgrade(w, r, nil)
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// // ensure connection close when function returns
// 	// defer ws.Close()
// 	// clients[ws] = true

// 	// for {
// 	// 	var msg Message
// 	// 	// Read in a new message as JSON and map it to a Message object
// 	// 	err := ws.ReadJSON(&msg)
// 	// 	if err != nil {
// 	// 		log.Printf("error: %v", err)
// 	// 		delete(clients, ws)
// 	// 		break
// 	// 	}
// 	// 	// send the new message to the broadcast channel
// 	// 	broadcast <- msg
// 	// }
// }
