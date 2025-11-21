#!/bin/bash

# Kill any existing processes on ports 8080 and 5173
fuser -k 8080/tcp
fuser -k 5173/tcp

# Start Backend (which now serves frontend)
echo "Starting Backend..."
cd backend
go run main.go &
BACKEND_PID=$!
cd ..

# Start Simulator (Optional)
echo "Starting Simulator..."
cd backend
sleep 2 # Wait for backend to start
go run simulator/main.go -id esp32-demo-1 &
SIMULATOR_PID=$!
cd ..

echo "Services started!"
echo "Backend PID: $BACKEND_PID"
echo "Simulator PID: $SIMULATOR_PID"
echo ""
echo "Access the app at http://localhost:8080"
echo "Press Ctrl+C to stop all services"

trap "kill $BACKEND_PID $SIMULATOR_PID; exit" INT

wait
