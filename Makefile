# Variables for Go compiler and flags
GO = go
GOFLAGS = -v

# Directories
SRCDIR1 = ./holepunchudptunnel
SRCDIR2 = ./udptunnel
BINDIR = bin

.PHONY: all clean make-prod

all: prepare-cn UDPTunnel HolePunch-UDPTunnel

UDPTunnel:
	cd $(SRCDIR2) && $(GO) build $(GOFLAGS) -o ../$(SRCDIR1)/tunnelman .

HolePunch-UDPTunnel:
	cd $(SRCDIR1) && $(GO) build $(GOFLAGS) -o ../$(BINDIR)/HolePunch-UDPTunnel .

prepare:
	cd $(SRCDIR1) && go work use .
	cd $(SRCDIR1) && go work use ./natholepunch
	cd $(SRCDIR1) && go work use ./tui
	cd $(SRCDIR1) && go work use ./tunnelman
	cd $(SRCDIR1) && go get
	cd $(SRCDIR2) && go get

prepare-cn:
	cd $(SRCDIR1) && go work use .
	cd $(SRCDIR1) && go work use ./natholepunch
	cd $(SRCDIR1) && go work use ./tui
	cd $(SRCDIR1) && go work use ./tunnelman
	export GOPROXY=https://goproxy.cn && export GOSUMDB=sum.golang.org && cd $(SRCDIR1) && go get
	export GOPROXY=https://goproxy.cn && export GOSUMDB=sum.golang.org && cd $(SRCDIR2) && go get

make-prod: GOFLAGS += -ldflags "-s -w" -trimpath
make-prod: all

clean:
	rm -rf $(BINDIR)
	rm -rf $(SRCDIR1)/tunnelman/udptunnel

# run-server:
# 	@echo "Running server..."
# 	$(BINDIR)/HolePunch-UDPTunnel --server $(ARGS)

# run-client:
# 	@echo "Running client..."
# 	$(BINDIR)/HolePunch-UDPTunnel $(ARGS)
