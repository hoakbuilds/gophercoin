PKG := github.com/murlokito/gophercoin

GOBUILD := go build
GOINSTALL := GO111MODULE=on go install -v

MKDIRBIN := mkdir bin
CHECKDIRBIN := bin/

# ============
# INSTALLATION
# ============

genproto:
	@$(call print, "Generating proto files.")
	$(GCRPRCPROTO)

build: | $(CHECKDIRBIN)
	@$(call print, "Building debug gophercoin.")
	$(GOBUILD) -o bin/gophercoind $(PKG)/cmd/gophercoind

install:
	@$(call print, "Installing gophercoin daemon.")
	$(GOINSTALL) $(PKG)/cmd/gophercoind

all:
	@$(call print, "Generating proto files.")
	@$(call print, "Installing gophercoin daemon.")
	$(GOINSTALL) $(PKG)/cmd/gophercoind

$(CHECKDIRBIN):
	@$(call print, "Folder $(CHECKDIRBIN) does not exist.")
	$(MKDIRBIN)

clear:
	rm go.*
	rm *.dat
	rm *.db

wipews:
	rm *.dat
	
wipedb:
	rm *.db