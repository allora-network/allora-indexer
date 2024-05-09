BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
BINARY_NAME := pump
# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

# BUILD_FLAGS := -ldflags '$(ldflags)'
BUILDDIR ?= $(CURDIR)

###########
# Install #
###########

# all: install

# install:
# 	# @echo "--> ensure dependencies have not been modified"
# 	# @go mod verify
# 	# @go mod tidy
# 	# @echo "--> installing allorad"
# 	@go install $(BUILD_FLAGS) -mod=readonly ./cmd/allorad

build:
	GOWORK=off go build -mod=readonly -o $(BUILDDIR)/${BINARY_NAME} .

run:
	go build -o $(BUILDDIR)/${BINARY_NAME}  .
	./${BINARY_NAME}

clean:
	go clean
	rm $(BUILDDIR)/${BINARY_NAME}
