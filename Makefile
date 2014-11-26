WORKDIR := $(CURDIR)/.workdir
WORKDIR_LNK := $(WORKDIR)/src
GIT := github.com
EXTERNAL_DEPENDENCIES := $(GIT)/tholowka/testing/assertions
EXTERNAL_DEPENDENCY_DIRS := $(addprefix $(CURDIR)/, $(EXTERNAL_DEPENDENCIES))

PACKAGES := events events/results

.PHONY: a-quick-build a-unit-test-check

######### CONFIGURATION #####################
#############################################
# This task allows the code to be linked, tested via Go commands. 
# Using a src directory interferes with go get behaviour
# when the project is used elsewhere.
# For this to work, GOPATH needs to be modified in the scope of Make.
available: $(WORKDIR_LNK) $(EXTERNAL_DEPENDENCY_DIRS)
$(WORKDIR_LNK):
	@mkdir -p $(WORKDIR)
	@ln -s $(CURDIR) $(WORKDIR_LNK)

$(EXTERNAL_DEPENDENCY_DIRS):
	@export GOPATH=$(WORKDIR) && go get $(EXTERNAL_DEPENDENCIES)



######## BUILD, UNIT-TEST, LINKING ##########
#############################################
clean:
	@rm -rf $(GIT)
	@rm -rf $(WORKDIR)/*.*

documentation:
	@export GOPATH=$(WORKDIR) && godoc $(PACKAGES)

a-quick-build: available
	@echo 'Running a quick build'
	@export GOPATH=$(WORKDIR) && go build $(PACKAGES)
	@echo 'Finished a quick build'
	@echo 'Compilation ended.'

a-unit-test-check: available 
	@echo 'Running unit-tests'
	@export GOPATH=$(WORKDIR) && go test $(PACKAGES)
	@echo 'Finished unit-tests'

a-benchmark-check: available 
	@echo 'Running benchmark'
	@export GOPATH=$(WORKDIR) && go test -bench=. $(PACKAGES)
	@echo 'Finished unit-tests'

a-build: available
	@echo 'Running a build (linking)'
	@export GOPATH=$(WORKDIR) && go clean $(PACKAGES) && go install $(PACKAGES)
	@echo 'Finished a build (linking)'
	@echo 'Linking ended.'
