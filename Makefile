OWNER = dnitsch
NAME := async-api-generator
GIT_TAG := "0.0.0"
VERSION := "v$(GIT_TAG)"
REVISION := "aaaabbbb"

LDFLAGS := -ldflags="-s -w -X \"github.com/$(OWNER)/$(NAME)/cmd.Version=$(VERSION)\" -X \"github.com/$(OWNER)/$(NAME)/cmd.Revision=$(REVISION)\" -extldflags -static"

install:
	go mod tidy
	go mod vendor
	
install_ci:
	go mod vendor

.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf dist/*
	rm -rf vendor/*
	mkdir -p dist

bingen:
	for os in darwin linux windows; do \
		GOOS=$$os CGO_ENABLED=0 go build -mod=readonly -buildvcs=false $(LDFLAGS) -o dist/$(NAME)-$$os ./cmd; \
	done

build: clean install bingen

build_ci: clean install_ci bingen

cross-build: bingen

tag: 
	git tag -a $(VERSION) -m "ci tag release" $(REVISION)
	git push origin $(VERSION)

release:
	OWNER=$(OWNER) NAME=$(NAME) PAT=$(PAT) VERSION=$(VERSION) . hack/release.sh 

test_prereq: 
	mkdir -p .coverage
	go install github.com/jstemmer/go-junit-report/v2@latest && \
	go install github.com/axw/gocov/gocov@latest && \
	go install github.com/AlekSi/gocov-xml@latest

test: test_prereq
	go test ./... -v -mod=readonly -race -coverprofile=.coverage/out | go-junit-report > .coverage/report-junit.xml && \
	gocov convert .coverage/out | gocov-xml > .coverage/report-cobertura.xml

coverage: test
	go tool cover -html=.coverage/out
