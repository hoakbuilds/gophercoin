PKG := github.com/murlokito/gophercoin

GOBUILD := go build
GOINSTALL := GO111MODULE=on go install -v
MKDIRBUiLD := mkdir build

GCRPRCPROTO := protoc -I gcd/gcrpc/api/ \
				-I${GOPATH}/src \
				--go_out=plugins=grpc:api \
				gcd/gcrpc/api/api.proto

# ============
# INSTALLATION
# ============

genproto:
	$(GCRPRCPROTO)


build:
	$(MKDIRBUiLD)
	@$(call print, "Building debug gophercoin and glcli.")
	$(GOBUILD) -o build/gophercoin gophercoin
	$(GOBUILD) -o gc-bxp $(PKG)/cmd/gcbxp

install:
	@$(call print, "Installing gophercoin and glcli.")
	$(GOINSTALL) $(PKG)
	$(GOINSTALL) $(PKG)/cmd/gcbxp

clear:
	rm go.*
	rm *.dat
	rm *.db