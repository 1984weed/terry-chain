package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// Message is a message
type Message struct {
	Type int         `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		fmt.Printf("Got message: %#v\n", message)
		fmt.Printf("err: %#v\n", err)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		switch message.Type {
		case queryLatest:
			c.sendMesssage(Message{
				Type: responseBlockchain,
				Data: GetLatestBlock(),
			})
		case queryAll:
			c.sendMesssage(Message{
				Type: responseBlockchain,
				Data: GetBlockchain(),
			})
		case responseBlockchain:
			if message.Data == nil {
				break
			}
		}

		c.handleBlockchainResponse(message.Data.([]Block))
	}
}
func responseLatestMsg() Message {
	return Message{
		Type: responseBlockchain,
		Data: GetLatestBlock(),
	}
}

func queryAllMsg() Message {
	return Message{
		Type: queryAll,
		Data: nil,
	}

}
func (c *Client) sendMesssage(message Message) {
	byte, _ := json.Marshal(message)
	c.send <- byte
}
func (c *Client) broadcast(message Message) {
	byte, _ := json.Marshal(message)
	c.hub.broadcast <- byte
}

func (c *Client) handleBlockchainResponse(receivedBlocks []Block) {
	if len(receivedBlocks) == 0 {
		fmt.Println("received block chain size of 0")
		return
	}
	latestBlockReceived := receivedBlocks[len(receivedBlocks)-1]

	latestBlockHeld := GetLatestBlock()

	if latestBlockReceived.Index > latestBlockHeld.Index {
		fmt.Println(fmt.Sprintf("blockchain possibly behind. We got: %#v Peer got: %#v", latestBlockReceived.Index, latestBlockReceived.Index))
		if latestBlockHeld.Hash == latestBlockReceived.PreviousHash {
			if addBlockToChain(latestBlockReceived) {
				c.broadcast(responseLatestMsg())
			}
		} else if len(receivedBlocks) == 1 {
			fmt.Println("We have to query the chain from our peer")
			c.broadcast(queryAllMsg())
		} else {
			fmt.Println("Received blockchain is longer than current blockchain")
			ReplaceChain(receivedBlocks)
		}
	} else {
		fmt.Println("received blockchain is not longer than received blockchain. Do nothing")
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

const (
	queryLatest        = 0
	queryAll           = 1
	responseBlockchain = 2
)

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
