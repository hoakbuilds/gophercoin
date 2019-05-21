PKG := github.com/murlokito/gophercoin

GOBUILD := go build
GOINSTALL := GO111MODULE=on go install -v
MKDIRBUiLD := mkdir bin

GCRPRCPROTO := protoc -I gcd/gcrpc/ \
				-I${GOPATH}/src \
				--go_out=plugins=grpc:gcd/gcrpc/ \
				gcd/gcrpc/api.proto

# ============
# INSTALLATION
# ============

genproto:
	@$(call print, "Generating proto files.")
	$(GCRPRCPROTO)

build:
	$(MKDIRBUiLD)
	@$(call print, "Building debug gophercoin and glcli.")
	$(GOBUILD) -o bin/gcd $(PKG)/cmd/gcd
	$(GOBUILD) -o bin/gccli $(PKG)/cmd/gccli

install: 
	@$(call print, "Installing gophercoin daemon (gcd) and gophercoin cli (gccli).")
	$(GOINSTALL) $(PKG)/cmd/gcd
	$(GOINSTALL) $(PKG)/cmd/gccli

all:
	@$(call print, "Generating proto files.")
	$(GCRPRCPROTO)
	@$(call print, "Installing gophercoin daemon (gcd) and gophercoin cli (gccli).")
	$(GOINSTALL) $(PKG)/cmd/gcd
	$(GOINSTALL) $(PKG)/cmd/gccli

clear:
	rm go.*
	rm *.dat
	rm *.db

wipews:
	rm *.dat
	
wipedb:
	rm *.db