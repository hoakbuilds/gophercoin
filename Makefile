PKG := github.com/murlokito/gophercoin

GOBUILD := go build
GOINSTALL := GO111MODULE=on go install -v
MKDIRBUiLD := mkdir build

<<<<<<< HEAD
GCRPRCPROTO := protoc -I gcd/gcrpc/ \
				-I${GOPATH}/src \
				--go_out=plugins=grpc:gcd/gcrpc/ \
				gcd/gcrpc/api.proto
=======
GCRPRCPROTO := protoc -I gcd/gcrpc/api/ \
				-I${GOPATH}/src \
				--go_out=plugins=grpc:api \
				gcd/gcrpc/api/api.proto
>>>>>>> d3233990347c6be6c9d1316dbc6bc74557aa1242

# ============
# INSTALLATION
# ============

genproto:
	$(GCRPRCPROTO)


build:
	$(MKDIRBUiLD)
	@$(call print, "Building debug gophercoin and glcli.")
<<<<<<< HEAD
	$(GOINSTALL) $(PKG)/cmd/gcd
	$(GOINSTALL) $(PKG)/cmd/gccli

install:
	@$(call print, "Installing gophercoin and glcli.")
	$(GOINSTALL) $(PKG)/cmd/gcd
	$(GOINSTALL) $(PKG)/cmd/gccli
=======
	$(GOBUILD) -o build/gophercoin gophercoin
	$(GOBUILD) -o gc-bxp $(PKG)/cmd/gcbxp

install:
	@$(call print, "Installing gophercoin and glcli.")
	$(GOINSTALL) $(PKG)
	$(GOINSTALL) $(PKG)/cmd/gcbxp
>>>>>>> d3233990347c6be6c9d1316dbc6bc74557aa1242

clear:
	rm go.*
	rm *.dat
	rm *.db