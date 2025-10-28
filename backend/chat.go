package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"github.com/coder/websocket"
)

// Message struct
type Message struct {
	User string `json:"user"`
	Text string `json:"text"`
	Time string `json:"time"`
}

type subscriber struct {
	name      string
	msgs      chan []byte
	closeSlow func()
}

type chatServer struct {
	subscriberMessageBuffer int
	publishLimiter          *rate.Limiter
	logf                    func(f string, v ...any)
	serveMux                http.ServeMux
	subscribersMu           sync.Mutex
	subscribers             map[*subscriber]struct{}
}

func (cs *chatServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cs.serveMux.ServeHTTP(w, r)
}

func (cs *chatServer) addSubscriber(s *subscriber) {
	cs.subscribersMu.Lock()
	cs.subscribers[s] = struct{}{}
	cs.subscribersMu.Unlock()
}

func (cs *chatServer) deleteSubscriber(s *subscriber) {
	cs.subscribersMu.Lock()
	delete(cs.subscribers, s)
	cs.subscribersMu.Unlock()
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.Write(ctx, websocket.MessageText, msg)
}

func (cs *chatServer) publish(msg Message) {
	data, _ := json.Marshal(msg)
	cs.subscribersMu.Lock()
	defer cs.subscribersMu.Unlock()
	cs.publishLimiter.Wait(context.Background())

	for s := range cs.subscribers {
		select {
		case s.msgs <- data:
		default:
			go s.closeSlow()
		}
	}
}

func (cs *chatServer) publishBytes(data []byte) {
	cs.subscribersMu.Lock()
	defer cs.subscribersMu.Unlock()
	cs.publishLimiter.Wait(context.Background())

	for s := range cs.subscribers {
		select {
		case s.msgs <- data: // forward raw bytes
		default:
			go s.closeSlow()
		}
	}
}


func (cs *chatServer) publishSystemMessage(text string) {
	msg := Message{
		User: "SYSTEM",
		Text: text,
		Time: time.Now().Format("15:04:05"),
	}
	cs.publish(msg)
}

func (cs *chatServer) subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Anonymous"
	}

	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool

	s := &subscriber{
		name: name,
		msgs: make(chan []byte, cs.subscriberMessageBuffer),
		closeSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "connection too slow")
			}
		},
	}

	cs.addSubscriber(s)
	defer func() {
		cs.deleteSubscriber(s)
		cs.publishSystemMessage(fmt.Sprintf("👋 %s left the chat", name))
	}()

	// Accept WebSocket with compression enabled
	c2, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionContextTakeover,
	})
	if err != nil {
		return err
	}

	mu.Lock()
	if closed {
		mu.Unlock()
		return nil
	}
	c = c2
	mu.Unlock()
	defer c.Close(websocket.StatusNormalClosure, "bye")

	cs.publishSystemMessage(fmt.Sprintf("💬 %s joined the chat", name))
	// ctx = c.CloseRead(ctx)

	// Reader goroutine: handle messages from client
	go func() {
		for {
			typ, msgBytes, err := c.Read(ctx)
			if err != nil {
				return
			}
			if typ != websocket.MessageText {
				continue
			}

			cs.publishBytes(msgBytes)
		}
	}()

	// Writer loop: send messages to client
	for {
		select {
		case data := <-s.msgs:
			if err := writeTimeout(ctx, 5*time.Second, c, data); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (cs *chatServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	err := cs.subscribe(r.Context(), w, r)
	if err != nil {
		cs.logf("%v", err)
	}
}

func newChatServer() *chatServer {
	cs := &chatServer{
		subscriberMessageBuffer: 16,
		logf:                    log.Printf,
		subscribers:             make(map[*subscriber]struct{}),
		publishLimiter:          rate.NewLimiter(rate.Every(100*time.Millisecond), 8),
	}
	cs.serveMux.HandleFunc("/subscribe", cs.subscribeHandler)
	return cs
}
