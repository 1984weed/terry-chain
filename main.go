package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	// http.HandleFunc("/", handler)
	// const app = express();
	// app.use(bodyParser.json());

	http.HandleFunc("/blocks", blocksHandler)
	http.HandleFunc("/mineBlock", blocksHandler)
	http.HandleFunc("/peers", blocksHandler)
	http.HandleFunc("/addPeer", blocksHandler)

	http.ListenAndServe(":8080", nil)
}

func blocksHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(GetBlockChain())
}
