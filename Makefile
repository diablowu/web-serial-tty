.PHONY: all build clean frontend backend

BINARY_NAME=web-serial-tty
BUILD_DIR=build

all: build

build: frontend backend

frontend:
	@echo "Building frontend..."
	cd frontend && npm install && npm run build
	@echo "Copying frontend assets to backend..."
	rm -rf backend/dist
	cp -r frontend/dist backend/dist

backend:
	@echo "Building backend..."
	cd backend && go mod tidy && go build -o ../$(BUILD_DIR)/$(BINARY_NAME) main.go

clean:
	@echo "Cleaning..."
	rm -rf backend/dist
	rm -rf $(BUILD_DIR)
	rm -rf frontend/dist
