package gmap

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/websocket"
	. "github.com/peak6/logger"
	"net"
	"net/http"
	"os"
	"sync"
)

const cluster_path = "/_cluster_ws"
const client_path = "/ws"

type NodeInfo struct {
	Name string
	Host string
	URL  string
	Addr string
	Pid  int
}

func (n *NodeInfo) String() string {
	return n.Name
}

type connection struct {
	ws       *websocket.Conn
	sendChan chan interface{}
	name     string
	node     *NodeInfo
}

var MyNode NodeInfo
var clusterWait sync.WaitGroup

func init() {
	var err error
	MyNode.Host, err = os.Hostname()
	if err != nil {
		panic(err)
	}
	gob.RegisterName("NodeInfo", NodeInfo{})
}

func Wait() {
	clusterWait.Wait()
}

func Listen(listen string) error {
	return ListenAndJoin(listen, nil)
}
func ListenAndJoin(listen string, join []string) error {
	defer clusterWait.Done()
	http.HandleFunc("/", serveRoot)
	http.HandleFunc(client_path, clientWs)
	http.HandleFunc(cluster_path, clusterWs)
	clusterWait.Add(1)
	lsock, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	defer lsock.Close()

	sockAddr, ok := lsock.Addr().(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("Expected tcp addr, but got: %v", lsock.Addr())
	}

	if listen[0] == ':' {
		MyNode.Addr = fmt.Sprintf("%s:%d", MyNode.Host, sockAddr.Port)
	} else {
		MyNode.Addr = sockAddr.String()
	}
	MyNode.Pid = os.Getpid()
	if MyNode.Name == "" {
		MyNode.Name = fmt.Sprintf("%s/%d", MyNode.Addr, MyNode.Pid)
	}
	MyNode.URL = "ws://" + MyNode.Addr + cluster_path
	Linfo.Printf("Node: %s, listening on: %s, URL: %s", MyNode.Name, MyNode.Addr, MyNode.URL)
	MyStore.PutStatic("/node", MyNode)
	for _, j := range join {
		err = Join(j)
		if err != nil {
			return err
		}
	}

	return http.Serve(lsock, nil)

}

func serveRoot(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Please connect on: /ws", 500)
}

func (c connection) String() string {
	return c.name
}

func NewConnection(ws *websocket.Conn) (*connection, error) {
	var err error
	ret := &connection{ws: ws}

	err = ws.WriteJSON(&MyNode)
	if err == nil {
		err = ws.ReadJSON(&ret.node)
	}

	if err != nil {
		return nil, err
	}

	ret.name = fmt.Sprintf("%s -> %s (%s)", ws.LocalAddr(), ws.RemoteAddr(), ret.node)
	ret.sendChan = make(chan interface{}, 1)
	go ret.sendLoop()
	return ret, nil
}

func initWs(w http.ResponseWriter, r *http.Request) *connection {
	if r.Method != "GET" {
		return wserr(w, r, "Method "+r.Method+" not allowed", 405)
	}
	// if r.Header.Get("Origin") != "http://"+r.Host {
	// 	http.Error(w, "Origin not allowed", 403)
	// 	return nil
	// }
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		return wserr(w, r, "Not a websocket handshake", 400)
	} else if err != nil {
		return wserr(w, r, "Failed to setup websocket", 500)
	}

	ret, err := NewConnection(ws)
	if err != nil {
		return wserr(w, r, err.Error(), 500)
	}
	return ret
}

func (c *connection) Close() {
	c.ws.Close()
	close(c.sendChan)
	MyStore.RemoveForNode(c.node)
}

func wserr(w http.ResponseWriter, r *http.Request, error string, code int) *connection {
	http.Error(w, error, code)
	Lerr.Println("Failed to establish socket with:", r.RemoteAddr, "response code:", code, "message:", error)
	return nil
}

func clusterWs(w http.ResponseWriter, r *http.Request) {
	c := initWs(w, r)
	if c == nil {
		return
	}
	c.clusterLoop()
}

func clientWs(w http.ResponseWriter, r *http.Request) {
	c := initWs(w, r)
	if c == nil {
		return
	}
	c.clientLoop()
}

func (c *connection) sendLoop() {
	buff := bytes.NewBuffer(nil)
	encoder := gob.NewEncoder(buff)
	for msg := range c.sendChan {
		buff.Reset()
		err := encoder.Encode(msg)
		// Linfo.Println(c, "Sending:", msg)
		if err != nil {
			Lerr.Println(c, "failed to encode", msg, err)
		} else {
			c.ws.WriteMessage(websocket.BinaryMessage, buff.Bytes())
			Linfo.Println(c, "sent", msg)
		}
	}
	Linfo.Println(c, "exiting send loop")
}

func (c *connection) send(msg interface{}) {
	c.sendChan <- msg
}

func (c *connection) syncMyEnries() {
	var msg Sync
	msg.Action = "sync"
	msg.Store = MyStore.GetMyEntries()
	c.send(msg)
	// Linfo.Println("Sync:", spew.Sdump(msg))
}

func (c *connection) clusterLoop() {
	defer c.Close()
	clusterWait.Add(1)
	defer clusterWait.Done()
	Linfo.Println("Connected to cluster node:", c)
	buff := bytes.NewBuffer(nil)
	decoder := gob.NewDecoder(buff)
	c.syncMyEnries()
	for {
		msgType, data, err := c.ws.ReadMessage()
		if err != nil {
			Lerr.Println(c, "Exiting due to", err)
			return
		}
		switch msgType {
		case websocket.TextMessage:
			Linfo.Println(c, " received unexpected text message", string(data))
			return
		case websocket.BinaryMessage:
			buff.Reset()
			buff.Write(data) // panic's on fail
			var cmd Sync
			err = decoder.Decode(&cmd)
			if err != nil {
				Lerr.Printf("%s failed to parse: %s\n%s", c, err, hex.Dump(data))
				return
			}
			err = processSync(&cmd)
			if err != nil {
				Lerr.Println("Failed to process:", cmd)
			}
		default:
			Lerr.Printf("%s received unknown message type: %d", c, msgType, hex.Dump(data))
			return
		}
	}
}

func processSync(cmd *Sync) error {
	switch cmd.Action {
	case "sync":
		MyStore.AddAll(cmd.Store)
	default:
		return fmt.Errorf("Unknown action: %s", cmd.Action)
	}
	return nil
}

func (c *connection) clientLoop() {
	defer c.ws.Close()
	clusterWait.Add(1)
	defer clusterWait.Done()
	Linfo.Println("Servicing:", c.ws.RemoteAddr(), "->", c.ws.LocalAddr())
	for {
		msgType, p, err := c.ws.ReadMessage()
		if err != nil {
			Lerr.Println("Exiting due read to:", err)
			return
		}
		println("Echoing", msgType, string(p))
		if err = c.ws.WriteMessage(msgType, p); err != nil {
			Lerr.Println("Exiting due to write error:", err)
			return
		}
	}
}

func Join(url string) error {
	if url[:4] != "ws://" {
		url = "ws://" + url + cluster_path
	}
	ws, resp, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return fmt.Errorf("Failed to connect: %s, %d", resp.Status, resp.StatusCode)
	}
	newCon, err := NewConnection(ws)
	if err != nil {
		return err
	}
	go newCon.clusterLoop()
	return nil
}
