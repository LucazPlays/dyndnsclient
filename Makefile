.PHONY: all build build-linux install uninstall clean

BINARY_NAME=dyndns-client

all: build

build:
	go build -o $(BINARY_NAME) .

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux .

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
	rm -f $(BINARY_NAME)-linux
