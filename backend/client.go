package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	"github.com/coder/websocket"
)

func RunClient() {
	ctx := context.Background()

	// Ask for name
	fmt.Print("Enter your name: ")
	var name string
	fmt.Scanln(&name)
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Anonymous"
	}

	url := fmt.Sprintf("ws://localhost:8080/subscribe?name=%s", name)
	c, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{
		CompressionMode: websocket.CompressionContextTakeover,
	})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("✅ Connected to chat server as", name)

	// Function to send SYSTEM leave message
	sendLeaveMessage := func() {
		leaveMsg := Message{
			User: "SYSTEM",
			Text: fmt.Sprintf("%s left the chat", name),
			Time: time.Now().Format("15:04:05"),
		}
		data, _ := json.Marshal(leaveMsg)
		_ = c.Write(ctx, websocket.MessageText, data)
		c.Close(websocket.StatusNormalClosure, "Client closed")
		fmt.Println("\nExiting chat...")
	}

	// Ctrl+C handler
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		sendLeaveMessage()
		os.Exit(0)
	}()

	// Readline prompt
	rl, err := readline.New("\033[1;36mType message: \033[0m") // cyan prompt
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()

	// Reader goroutine for incoming messages
	go func() {
		for {
			typ, data, err := c.Read(ctx)
			if err != nil {
				log.Println("Disconnected:", err)
				return
			}
			if typ != websocket.MessageText {
				continue
			}

			var msg Message
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}

			timeStr := fmt.Sprintf("\033[97m[%s]\033[0m", msg.Time) // white timestamp
			var displayName string
			if msg.User == name {
				displayName = "Me"
			} else {
				displayName = msg.User
			}

			var text string
			switch msg.User {
			case "SYSTEM":
				text = fmt.Sprintf("%s \033[1;31m[SYSTEM]\033[0m %s", timeStr, msg.Text)
			case name:
				text = fmt.Sprintf("%s \033[1;32m\033[1m%s\033[0m: %s", timeStr, displayName, msg.Text)
			default:
				text = fmt.Sprintf("%s \033[1;34m\033[1m%s\033[0m: %s", timeStr, displayName, msg.Text)
			}

			fmt.Print("\033[2K\r") // clear line
			fmt.Println(text)
			rl.Refresh()
		}
	}()

	// Input loop with persistent prompt
	for {
		line, err := rl.Readline()
		if err != nil {
			// Exit on Ctrl+D or error
			sendLeaveMessage()
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.ToLower(line) == "exit" {
			sendLeaveMessage()
			break
		}

		msg := Message{
			User: name,
			Text: line,
			Time: time.Now().Format("15:04:05"),
		}

		data, _ := json.Marshal(msg)
		if err := c.Write(ctx, websocket.MessageText, data); err != nil {
			log.Println("write error:", err)
			sendLeaveMessage()
			break
		}
	}
}
