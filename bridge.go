package bridge

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

const BufferBytes = 8 << 10

type TCPListener struct {
	Address       string
	RemoteAddress string
}

func NewTCPListener() *TCPListener {
	return &TCPListener{
		Address:       "localhost:28080",
		RemoteAddress: "localhost:28081",
	}
}

func (l *TCPListener) Listen() {
	listener, err := net.Listen("tcp", l.Address)
	if err != nil {
		log.Println("Failed to start listening tcp.")
		return
	}
	log.Printf("TCP server is running at %s\n", l.Address)

	for {
		conn, err := listener.Accept()
		log.Printf("Accept: %v", conn.RemoteAddr().String())
		if err != nil {
			log.Println("Failed to connect TCP peer.")
			continue
		}
		wsReader, tcpWriter := io.Pipe()
		tcpReader, wsWriter := io.Pipe()
		tcpConn := &TCPConn{conn}
		go tcpConn.Handle(tcpReader, tcpWriter)
		err = l.TranslateWebsocket(wsReader, wsWriter)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

// TranslateWebsocket makes Websocket connection with remote server
// and call WebsocketConn.Handle.
func (l *TCPListener) TranslateWebsocket(r io.Reader, w io.WriteCloser) error {
	url := url.URL{Scheme: "ws", Host: l.RemoteAddress, Path: "/"}
	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Println("Failed to connect Websocket peer.")
		return err
	}
	wsConn := WebsocketConn{conn}
	go wsConn.Handle(r, w)
	return nil
}

type HTTPListener struct {
	Address       string
	RemoteAddress string
}

func NewHTTPListener() *HTTPListener {
	return &HTTPListener{
		Address:       "localhost:28081",
		RemoteAddress: "localhost:21080",
	}
}
func (l *HTTPListener) Listen() {
	upgrader := websocket.Upgrader{}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade websocket.")
			return
		}
		log.Printf("Upgrade: %s", conn.RemoteAddr().String())
		wsConn := WebsocketConn{conn}
		wsReader, tcpWriter := io.Pipe()
		tcpReader, wsWriter := io.Pipe()
		go wsConn.Handle(wsReader, wsWriter)
		err = l.TranslateTCP(tcpReader, tcpWriter)
		if err != nil {
			log.Println(err.Error())
		}
	})
	log.Printf("Websocket server is running at %s\n", l.Address)
	http.ListenAndServe(l.Address, nil)
}

// TranslateTCP makes TCP connection with remote host and call TCPConn.Handle.
func (l *HTTPListener) TranslateTCP(r io.Reader, w io.WriteCloser) error {
	conn, err := net.Dial("tcp", l.RemoteAddress)
	if err != nil {
		log.Println("Failed to connect tcp peer.")
		return err
	}
	tcpConn := TCPConn{conn}
	go tcpConn.Handle(r, w)
	return nil
}

// TCPConn is wrapper of net.Conn.
type TCPConn struct {
	c net.Conn
}

// Handle reads from TCP connection and write it to w.
// Also, reads from r and write it to TCP connection.
func (t *TCPConn) Handle(r io.Reader, w io.WriteCloser) {
	defer t.c.Close()
	done1 := make(chan struct{})
	done2 := make(chan struct{})

	go func() {
		defer close(done1)
		io.Copy(t.c, r)
	}()
	go func() {
		defer close(done2)
		defer w.Close()
		io.Copy(w, t.c)
	}()

	// wait for done either
	select {
	case <-done1:
	case <-done2:
	}
	log.Printf("TCP closed: %s", t.c.RemoteAddr().String())
}

// WebsocketConn is wrapper of websocket.Conn.
type WebsocketConn struct {
	*websocket.Conn
}

// Handle reads from Websocket connection and write it to w.
// Also, reads from r and write it to Websocket connection.
func (ws *WebsocketConn) Handle(r io.Reader, w io.WriteCloser) {
	defer ws.Close()
	done1 := make(chan struct{})
	done2 := make(chan struct{})

	go func() {
		defer close(done1)
		for {
			b := make([]byte, BufferBytes)
			size, err1 := r.Read(b)
			err2 := ws.WriteMessage(websocket.BinaryMessage, b[:size])
			if err2 != nil {
				// log.Println("Error: write websocket")
				return
			}
			if err1 != nil {
				// log.Println("Error: translate tcp to websocket")
				return
			}
		}
	}()

	go func() {
		defer close(done2)
		defer w.Close()
		for {
			_, message, err1 := ws.ReadMessage()
			_, err2 := w.Write(message)
			if err2 != nil {
				// log.Println("Error: translate websocket to tcp")
				return
			}
			if err1 != nil {
				// log.Println("Error: read websocket")
				return
			}
		}
	}()

	// wait for done either
	select {
	case <-done1:
	case <-done2:
	}
	log.Printf("Websocket closed: %s", ws.RemoteAddr().String())
}
