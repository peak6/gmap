package main

import (
	"encoding/hex"
	"github.com/gorilla/websocket"
	. "github.com/peak6/logger"
	"net/http"
	"time"
)

type connection struct {
	ws *websocket.Conn
}

func main() {
	InitLogger()
	con, resp, err := websocket.DefaultDialer.Dial("ws://localhost:1234/ws", http.Header{})
	if err != nil {
		Lerr.Println("Exiting due to dial error:", err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		Lerr.Println("Exiting due to unexpected status code:", resp.Status, resp.StatusCode)
		return
	}

	c := connection{ws: con}
	con.SetPingHandler(pingHandler)
	con.SetPongHandler(pongHandler)
	go c.traceLoop()
	for i := 0; i < 10; i++ {
		switch i % 3 {
		case 0:
			err = c.ws.WriteMessage(websocket.TextMessage, []byte("Text message"))
		case 1:
			err = c.ws.WriteMessage(websocket.BinaryMessage, []byte("Binary mesasge"))
		case 2:
			err = c.ws.WriteMessage(websocket.PingMessage, []byte("pping"))
		}
		if err != nil {
			Lerr.Println("Failed to send:", err)
		}
	}
	time.Sleep(1 * time.Second)
	println("CLosing")
	c.ws.Close()
	println("closed")
	time.Sleep(1 * time.Second)
}

func pongHandler(msg string) error {
	Linfo.Println("Received pong", msg)
	return nil
}
func pingHandler(msg string) error {
	Linfo.Println("Received ping", msg)
	return nil
}

func (c *connection) traceLoop() {
	defer c.ws.Close()
	for {
		msgType, data, err := c.ws.ReadMessage()
		if err != nil {
			Lerr.Println("Exiting traceLoop due to read error:", err)
			return
		}
		switch msgType {
		case websocket.TextMessage:
			Linfo.Println("Received:", string(data))
		case websocket.BinaryMessage:
			Linfo.Println("Received:", hex.Dump(data))
		default:
			Lerr.Println("Unknown type:", msgType)
			return
		}
	}

}
