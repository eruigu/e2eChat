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

type OnlineUpdate struct{
	Type string `json:"type"` 
	Count int `json:"count"`
}

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

func (cs *chatServer) publishOnlineCount() {
	cs.subscribersMu.Lock()
	count := len(cs.subscribers)

	data, _ := json.Marshal(OnlineUpdate{
		Type:  "online_count",
		Count: count,
	})

	subs := make([]*subscriber, 0, len(cs.subscribers))
	for s := range cs.subscribers {
		subs = append(subs, s)
	}
	cs.subscribersMu.Unlock()

	for _, s := range subs {
		select {
		case s.msgs <- data:
		default:
			go s.closeSlow()
		}
	}
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
				_ = c.Close(websocket.StatusPolicyViolation, "connection too slow")
			}
		},
	}

	cs.addSubscriber(s)
	cs.publishOnlineCount()

	defer func() {
		cs.deleteSubscriber(s)
	cs.publishOnlineCount()
		cs.publishSystemMessage(fmt.Sprintf("👋 %s left the chat", name))
	}()

	c2, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns:  []string{"localhost:5173"},
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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer func() {
		_ = c.Close(websocket.StatusNormalClosure, "bye")
	}()

	cs.publishSystemMessage(fmt.Sprintf("💬 %s joined the chat", name))

	// Reader goroutine
	go func() {
		defer cancel()

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

	// Writer loop
	for {
		select {
		case data, ok := <-s.msgs:
			if !ok {
				return nil
			}
			if err := writeTimeout(ctx, 5*time.Second, c, data); err != nil {
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (cs *chatServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	if err := cs.subscribe(r.Context(), w, r); err != nil {
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
