# Web Serial TTY

A web-based serial TTY interface for ESP32 devices, featuring a Golang backend and React frontend. The frontend is embedded into the Golang binary for a single-file deployment.

## Features

*   **Web-based TTY**: Interact with your ESP32's serial console via a browser.
*   **WebSocket Transport**: Real-time bidirectional communication.
*   **Device Discovery**: Automatically lists connected devices.
*   **Hex/ASCII Modes**: Support for sending data in both ASCII and Hex formats.
*   **Embedded Frontend**: Single binary deployment.
*   **Cross-Platform**: Runs on Linux, Windows, and macOS.

## Building

### Prerequisites

*   Go 1.21+
*   Node.js 20+
*   Make

### Build Command

To build the complete project (frontend + backend):

```bash
make build
```

The output binary will be located in `build/web-serial-tty`.

## Running

### Basic Usage

```bash
./build/web-serial-tty
```

Access the interface at `http://localhost:8080`.

### Configuration

You can configure the bind address using the `-addr` flag:

```bash
./build/web-serial-tty -addr :9090
```

### Simulator

A simulator is included for testing without a real device:

```bash
go run backend/simulator/main.go -id esp32-sim-1
```

## API Endpoints

*   `GET /api/devices`: List connected devices.
*   `WS /ws/device?id=<device_id>`: WebSocket endpoint for ESP32 devices.
*   `WS /ws/client?device_id=<device_id>`: WebSocket endpoint for the web client.

## License

MIT
