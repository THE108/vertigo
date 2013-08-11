// testserver
package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Proxy struct {
	port   int
	name   string
	log    *os.File
	client *http.Client
	ch     chan string
}

func (h Proxy) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	sReq := fmt.Sprintf("Request<br>%s %s %s<br>Host: %s<br>", req.Method,
		req.RequestURI, req.Proto, req.Host)

	clientReq, err := http.NewRequest(req.Method, req.RequestURI, nil)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	for header, value := range req.Header {
		v := strings.Join(value, "")
		sReq += fmt.Sprintf("%s: %s<br>", header, v)
		clientReq.Header.Set(header, v)
	}

	clientResp, err := h.client.Do(clientReq)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	defer clientResp.Body.Close()

	for header, value := range clientResp.Header {
		v := strings.Join(value, "")
		resp.Header().Set(header, v)
	}

	resp.Header().Set("X-Proxy", "VertiGoProxy")

	io.Copy(resp, clientResp.Body)

	h.log.WriteString(sReq)
	h.log.Sync()

	fmt.Print(sReq)

	h.ch <- sReq
}

func (h Proxy) run() {
	fmt.Println("Server started")

	defer func() {
		if err := h.log.Close(); err != nil {
			fmt.Println("Error closing file: %s", err)
		}
	}()

	h.log.WriteString("Start logging\n")

	serv := &http.Server{
		Addr:           fmt.Sprintf(":%d", h.port),
		Handler:        h,
		ReadTimeout:    100 * time.Second,
		WriteTimeout:   100 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := serv.ListenAndServe(); err != nil {
		fmt.Println("Error: %s", err.Error())
	}
}

func NewProxy(port int, name string) *Proxy {
	f, err := os.OpenFile("/tmp/gohttp.log", os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println("Error open file: %s", err.Error())
		return nil
	}

	p := &Proxy{
		port:   port,
		name:   name,
		log:    f,
		client: &http.Client{},
		ch:     make(chan string, 256),
	}
	return p
}

type ProxyConf struct {
	Port int
	Name string
}

func WSHandler(ws *websocket.Conn) {
	fmt.Printf("Websocket connected from %s\n", ws.Request().RemoteAddr)
	defer func() {
		if err := ws.Close(); err != nil {
			fmt.Println("Websocket could not be closed", err.Error())
		}
	}()

	var err error
	var data ProxyConf

	if err = websocket.JSON.Receive(ws, &data); err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
	}
	fmt.Printf("Port: %d Name: %s\n", data.Port, data.Name)

	proxy := NewProxy(data.Port, data.Name)
	go proxy.run()

	for msg := range proxy.ch {
		fmt.Println("Message sent")
		if err := websocket.Message.Send(ws, msg); err != nil {
			fmt.Printf("Error: %s", err.Error())
			break
		}
	}

	fmt.Printf("Websocket disconnected from %s", ws.Request().RemoteAddr)
}

func main() {
	pwd, _ := os.Getwd()
	http.Handle("/ws", websocket.Handler(WSHandler))
	http.Handle("/", http.FileServer(http.Dir(pwd)))
	http.Handle("/static", http.FileServer(http.Dir(pwd+"/static")))

	if err := http.ListenAndServe(":8000", nil); err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
