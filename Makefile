PKG := github.com/murlokito/gophercoin

GOBUILD := go build
GOINSTALL := GO111MODULE=on go install -v
MKDIRBUiLD := mkdir build

# ============
# INSTALLATION
# ============

build:
	$(MKDIRBUiLD)
	@$(call print, "Building debug gophercoin and glcli.")
	$(GOBUILD) -o build/gophercoin gophercoin


install:
	@$(call print, "Installing gophercoin and glcli.")
	$(GOINSTALL) $(PKG)

clear:
	rm
	rm go.*