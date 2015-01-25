UNAME := $(shell uname)
WORKDIR := $(CURDIR)/.workdir
WORKDIR_LNK := $(WORKDIR)/src
GIT := github.com
EXTERNAL_DEPENDENCIES := $(GIT)/tholowka/testing/assertions
EXTERNAL_DEPENDENCY_DIRS := $(addprefix $(CURDIR)/, $(EXTERNAL_DEPENDENCIES))

ifeq ($(UNAME), Linux) 
	GO_BINARIES_NAME_FEDORA := go1.4.1.linux-amd64.tar.gz
	GO_BINARIES_DOWNLOADED := $(WORKDIR)/$(GO_BINARIES_NAME_FEDORA) 
	GO_DOWNLOAD_BINARY_URL := https://storage.googleapis.com/golang/$(GO_BINARIES_NAME_FEDORA)
endif
ifeq ($(UNAME), Darwin)
	GO_BINARIES_NAME_OSX := go1.4.1.darwin-amd64-osx10.8.tar.gz 
	GO_BINARIES_DOWNLOADED := $(WORKDIR)/$(GO_BINARIES_NAME_OSX) 
	GO_DOWNLOAD_BINARY_URL := https://storage.googleapis.com/golang/$(GO_BINARIES_NAME_OSX)
endif 
PACKAGES := events 
BUILD_PATH :=  $(GOPATH):$(WORKDIR)

.PHONY: go-available a-quick-build a-unit-test-check

#
# Go aspects 
#
.PHONY: go-available
.DELETE_ON_ERROR: $(GO_BINARIES_DOWNLOADED) $(GO_DIR)
GO_WGET_CMD := wget -O $(GO_BINARIES_DOWNLOADED) -nc $(GO_DOWNLOAD_BINARY_URL) 
GO_UNTARRED_DIR := $(basename $(GO_BINARIES_DOWNLOADED))
GO_DIR := $(WORKDIR)/go
GO := $(GO_DIR)/bin/go
GODOC := $(GO_DIR)/bin/godoc

go-available: $(WORKDIR) $(GO_BINARIES_DOWNLOADED) $(GO_DIR) 
	@echo 'Go downloaded and available'

$(GO_BINARIES_DOWNLOADED): 
	$(shell $(GO_WGET_CMD))

$(GO_DIR):
	@tar -xzvf $(GO_BINARIES_DOWNLOADED)
	@mv go $(WORKDIR)

######### CONFIGURATION #####################
#############################################
# This task allows the code to be linked, tested via Go commands. 
# Using a src directory interferes with go get behaviour
# when the project is used elsewhere.
# For this to work, GOPATH needs to be modified in the scope of Make.
.PHONY: available
.DELETE_ON_ERROR: $(WORKDIR) $(WORKDIR_LNK) $(EXTERNAL_DEPENDENCY_DIRS)
available: go-available $(WORKDIR) $(WORKDIR_LNK) $(EXTERNAL_DEPENDENCY_DIRS)
$(WORKDIR): 
	@mkdir -p $(WORKDIR)
$(WORKDIR_LNK):
	@ln -s $(CURDIR) $(WORKDIR_LNK)

$(EXTERNAL_DEPENDENCY_DIRS):
	@export GOPATH=$(WORKDIR) && export GOROOT=$(GO_DIR) && $(GO) get $(EXTERNAL_DEPENDENCIES)

######## BUILD, UNIT-TEST, LINKING ##########
#############################################
PHONY: clean documentation a-quick-build a-quick-test a-single-test a-benchmark-test a-parallel-benchmark-test a-build
clean: 
	@rm -rf $(GIT)
	@rm -rf $(WORKDIR)

documentation:
	@export GOPATH=$(WORKDIR) && export GOROOT=$(GO_DIR) && $(GODOC) $(PACKAGES)

a-quick-build: available
	@echo 'Running a quick build'
	@export GOPATH=$(BUILD_PATH) && export GOROOT=$(GO_DIR) && $(GO) build $(PACKAGES)
	@export GOPATH=$(WORKDIR) && go build $(PACKAGES)
	@echo 'Finished a quick build'
	@echo 'Compilation ended.'

a-quick-test: available 
	@echo 'Running unit-tests'
	@export GOPATH=$(BUILD_PATH) && export GOROOT=$(GO_DIR) && $(GO)  test $(PACKAGES)
	@echo 'Finished unit-tests'

a-single-test: available
	@echo 'Running unit-tests'
	@export GOPATH=$(BUILD_PATH) && export GOROOT=$(GO_DIR) && $(GO) test -run Test_And_WithMultipleTopics_And_That_It_DoesntWait* $(PACKAGES)
	@echo 'Finished unit-tests'

a-benchmark-check: available 
	@echo 'Running benchmark'
	@export GOPATH=$(BUILD_PATH) && export GOROOT=$(GO_DIR) && $(GO) test -bench=. -benchmem $(PACKAGES)
	@echo 'Finished unit-tests'

a-parallel-benchmark-check: available 
	@echo 'Running parallel benchmark'
	@export GOPATH=$(BUILD_PATH) && export GOROOT=$(GO_DIR) && $(GO) test -bench=Benchmark_Parallel_Topics -benchmem $(PACKAGES)
	@echo 'Finished unit-tests'

a-build: available
	@echo 'Running a build (linking)'
	@export GOPATH=$(BUILD_PATH) && export GOROOT=$(GO_DIR) && $(GO) install $(PACKAGES)
	@echo 'Finished a build (linking)'
	@echo 'Linking ended.'
