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
	hub := newHub()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "blockchain.html")
	})
	http.HandleFunc("/blocks", blocksHandler)
	http.HandleFunc("/mineBlock", mineBlock)
	http.HandleFunc("/peers", func(w http.ResponseWriter, r *http.Request) {
		response := make([]string, len(hub.clients))
		var i = 0
		for c, value := range hub.clients {
			if value {
				response[i] = c.conn.RemoteAddr().String()
				i++
			}
		}
		json.NewEncoder(w).Encode(response)
	})

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
