package main

import (
    "fmt"
    "log"
    "net/http"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}
type Client struct {
    conn *websocket.Conn
    send chan []byte
}
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}
func newHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan []byte),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}
func (h *Hub) run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
        case message := <-h.broadcast:
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
        }
    }
}
func (c *Client) readPump(hub *Hub) {
    defer func() {
        hub.unregister <- c
        c.conn.Close()
    }()
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            log.Println("read:", err)
            break
        }
        hub.broadcast <- message
    }
}
func (c *Client) writePump() {
    for message := range c.send {
        err := c.conn.WriteMessage(websocket.TextMessage, message)
        if err != nil {
            log.Println("write:", err)
            break
        }
    }
    c.conn.Close()
}
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("upgrade:", err)
        return
    }
    client := &Client{conn: conn, send: make(chan []byte, 256)}
    hub.register <- client

    go client.writePump()
    go client.readPump(hub)
}
func StartServer() {
    hub := newHub()
    go hub.run()

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        serveWs(hub, w, r)
    })

    addr := "localhost:8080"
    fmt.Printf("Server started at %s\n", addr)
    err := http.ListenAndServe(addr, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
