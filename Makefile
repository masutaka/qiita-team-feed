NAME := qiita-team-feed
SRCS := $(shell find . -type f -name '*.go' ! -name '*_test.go')
PACKAGES := $(shell go list ./...)

ifeq (Windows_NT, $(OS))
NAME := $(NAME).exe
endif

all: $(NAME)

# Install dependencies for development
.PHONY: deps
deps: dep
	dep ensure

.PHONY: dep
dep:
ifeq ($(shell command -v dep 2> /dev/null),)
	go get github.com/golang/dep/cmd/dep
endif

# Build binary
$(NAME): $(SRCS)
	go build -o $(NAME)

# Install binary to $GOPATH/bin
.PHONY: install
install:
	go install

# Clean binary
.PHONY: clean
clean:
	rm -f $(NAME)

# Test for development
.PHONY: test
test:
	go test -v $(PACKAGES)

# Test for CI
.PHONY: test-all
test-all: deps-test-all vet lint test

.PHONY: deps-test-all
deps-test-all: dep golint
	dep ensure

.PHONY: golint
golint:
ifeq ($(shell command -v golint 2> /dev/null),)
	go get github.com/golang/lint/golint
endif

.PHONY: vet
vet:
	go vet $(PACKAGES)

.PHONY: lint
lint:
	echo $(PACKAGES) | xargs -n1 golint -set_exit_status
