package main

import (
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

//go:embed dist/*
var content embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域，方便开发
	},
}

// Device represents a connected ESP32
type Device struct {
	ID   string
	Conn *websocket.Conn
	// 可以添加更多元数据，如 IP 等
}

// Hub maintains the set of active devices and broadcasts messages
type Hub struct {
	Devices      map[string]*Device
	DevicesMutex sync.RWMutex
	Clients      map[*websocket.Conn]string // Client Conn -> Device ID
	ClientsMutex sync.RWMutex
}

var hub = Hub{
	Devices: make(map[string]*Device),
	Clients: make(map[*websocket.Conn]string),
}

func main() {
	addr := flag.String("addr", ":8080", "http service address")
	flag.Parse()

	http.HandleFunc("/ws/device", handleDeviceWebsocket)
	http.HandleFunc("/ws/client", handleClientWebsocket)
	http.HandleFunc("/api/devices", handleListDevices)

	// Serve static files
	distFS, err := fs.Sub(content, "dist")
	if err != nil {
		log.Fatal(err)
	}
	fileServer := http.FileServer(http.FS(distFS))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if file exists in dist
		path := r.URL.Path
		if path == "/" {
			path = "index.html"
		} else {
			path = path[1:] // remove leading /
		}

		f, err := distFS.Open(path)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// If not found, serve index.html for SPA routing
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	log.Printf("Server started on %s", *addr)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// handleDeviceWebsocket handles websocket requests from ESP32
func handleDeviceWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	// 假设设备在连接 URL 中通过 query 参数传递 ID，或者第一条消息发送 ID
	// 这里简单起见，使用 query 参数 ?id=xxx
	deviceID := r.URL.Query().Get("id")
	if deviceID == "" {
		// 如果没有 ID，生成一个或者拒绝
		deviceID = conn.RemoteAddr().String()
	}

	device := &Device{
		ID:   deviceID,
		Conn: conn,
	}

	hub.DevicesMutex.Lock()
	hub.Devices[deviceID] = device
	hub.DevicesMutex.Unlock()

	log.Printf("Device connected: %s", deviceID)

	defer func() {
		hub.DevicesMutex.Lock()
		delete(hub.Devices, deviceID)
		hub.DevicesMutex.Unlock()
		conn.Close()
		log.Printf("Device disconnected: %s", deviceID)
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		// 收到设备发来的数据，转发给所有连接到该设备的客户端
		broadcastToClients(deviceID, messageType, message)
	}
}

// handleClientWebsocket handles websocket requests from Web TTY
func handleClientWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: device_id required"))
		conn.Close()
		return
	}

	hub.ClientsMutex.Lock()
	hub.Clients[conn] = deviceID
	hub.ClientsMutex.Unlock()

	log.Printf("Client connected to device: %s", deviceID)

	defer func() {
		hub.ClientsMutex.Lock()
		delete(hub.Clients, conn)
		hub.ClientsMutex.Unlock()
		conn.Close()
		log.Printf("Client disconnected from device: %s", deviceID)
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		// 收到客户端发来的数据，转发给对应的设备
		sendToDevice(deviceID, messageType, message)
	}
}

func broadcastToClients(deviceID string, messageType int, message []byte) {
	hub.ClientsMutex.RLock()
	defer hub.ClientsMutex.RUnlock()

	for clientConn, targetDeviceID := range hub.Clients {
		if targetDeviceID == deviceID {
			err := clientConn.WriteMessage(messageType, message)
			if err != nil {
				log.Printf("error writing to client: %v", err)
				clientConn.Close()
				// 注意：这里不能直接从 map 中删除，因为我们在遍历 map
				// 实际生产环境需要更健壮的处理，例如使用 channel
			}
		}
	}
}

func sendToDevice(deviceID string, messageType int, message []byte) {
	hub.DevicesMutex.RLock()
	device, ok := hub.Devices[deviceID]
	hub.DevicesMutex.RUnlock()

	if ok {
		err := device.Conn.WriteMessage(messageType, message)
		if err != nil {
			log.Printf("error writing to device: %v", err)
		}
	}
}

func handleListDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	hub.DevicesMutex.RLock()
	defer hub.DevicesMutex.RUnlock()

	var devices []string
	for id := range hub.Devices {
		devices = append(devices, id)
	}

	json.NewEncoder(w).Encode(devices)
}
