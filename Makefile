PKG := github.com/murlokito/gophercoin

GOBUILD := go build
GOINSTALL := GO111MODULE=on go install -v
MKDIRBUiLD := mkdir build

GCRPRCPROTO := protoc -I gcd/gcrpc/ \
				-I${GOPATH}/src \
				--go_out=plugins=grpc:gcd/gcrpc/ \
				gcd/gcrpc/api.proto

# ============
# INSTALLATION
# ============

genproto:
	$(GCRPRCPROTO)


build:
	$(MKDIRBUiLD)
	@$(call print, "Building debug gophercoin and glcli.")
	$(GOINSTALL) $(PKG)/cmd/gcd
	$(GOINSTALL) $(PKG)/cmd/gccli

install:
	@$(call print, "Installing gophercoin and glcli.")
	$(GOINSTALL) $(PKG)/cmd/gcd
	$(GOINSTALL) $(PKG)/cmd/gccli

clear:
	rm go.*
	rm *.dat
	rm *.db