.PHONY: all build build-linux-amd64 build-linux-arm64 hashes install uninstall clean

BINARY_NAME=dyndns-client

all: build

build:
	go build -o $(BINARY_NAME) ./cmd/dyndns

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 ./cmd/dyndns

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o $(BINARY_NAME)-linux-arm64 ./cmd/dyndns

hashes: build-linux-amd64 build-linux-arm64
	sha256sum $(BINARY_NAME)-linux-amd64 > $(BINARY_NAME)-linux-amd64.sha256
	sha256sum $(BINARY_NAME)-linux-arm64 > $(BINARY_NAME)-linux-arm64.sha256

install: build
	sudo cp $(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	sudo chown root:root /usr/local/bin/$(BINARY_NAME)

uninstall:
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	sudo systemctl stop $(BINARY_NAME) 2>/dev/null || true
	sudo systemctl disable $(BINARY_NAME) 2>/dev/null || true
	sudo rm -f /etc/systemd/system/$(BINARY_NAME).service
	sudo systemctl daemon-reload
	sudo rm -f /etc/dyndns-client.conf
	rm -f ~/.dyndns-client.addr

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux-amd64*
	rm -f $(BINARY_NAME)-linux-arm64*
