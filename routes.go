package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler = func(w http.ResponseWriter, r *http.Request)

func createRoutes() {
	r := mux.NewRouter()
	hub := newHub()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "blockchain.html")
	}).Methods("GET")
	r.HandleFunc("/blocks", blocksHandler).Methods("GET")
	r.HandleFunc("/blocks/:hash", blocksHandler).Methods("GET")

	r.HandleFunc("/mineBlock", mineBlock).Methods("POST")
	r.HandleFunc("/peers", getPeers(hub)).Methods("POST")

	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
}

func getPeers(hub *Hub) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		response := make([]string, len(hub.clients))
		var i = 0
		for c, value := range hub.clients {
			if value {
				response[i] = c.conn.RemoteAddr().String()
				i++
			}
		}
		json.NewEncoder(w).Encode(response)
	}
}

func blocksHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(GetBlockchain())
}
func getBlockByHashHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var resBlock *Block
	for _, block := range GetBlockchain() {
		if block.Hash == vars["hash"] {
			resBlock = &block
			break
		}
	}
	json.NewEncoder(w).Encode(resBlock)
}

func mineBlock(w http.ResponseWriter, r *http.Request) {
	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(w, "Error reading request body",
	// 		http.StatusInternalServerError)
	// }
	// results := string(body)
	json.NewEncoder(w).Encode(GenerateNextBlock())
}
