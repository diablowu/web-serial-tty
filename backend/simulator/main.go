package main

import (
	"bufio"
	"flag"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	id := flag.String("id", "esp32-sim-1", "Device ID")
	addr := flag.String("addr", "localhost:80", "Server address")
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws/device", RawQuery: "id=" + *id}
	log.Printf("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			if mt == websocket.BinaryMessage {
				log.Printf("recv (hex): %x", message)
				// Echo back as Binary
				err = c.WriteMessage(websocket.BinaryMessage, message)
				if err != nil {
					log.Println("write:", err)
					return
				}
			} else {
				log.Printf("recv: %s", message)
				// Echo back with a prefix to distinguish
				err = c.WriteMessage(websocket.TextMessage, []byte("ESP32: "+string(message)))
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Read from stdin to simulate serial output
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			err := c.WriteMessage(websocket.TextMessage, []byte(text+"\r\n"))
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("Log: "+t.String()+"\r\n"))
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
}
