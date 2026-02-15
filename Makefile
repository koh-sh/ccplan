.PHONY: test fmt cov tidy run lint fix build

COVFILE = coverage.out
COVHTML = cover.html
GITHUB_REPOSITORY = koh-sh/ccplan

test:
	go test ./... -json | go tool tparse -all

fmt:
	go tool gofumpt -l -w .

cov:
	go test -cover ./... -coverprofile=$(COVFILE)
	go tool cover -html=$(COVFILE) -o $(COVHTML)
	CI=1 GITHUB_REPOSITORY=$(GITHUB_REPOSITORY) octocov
	rm $(COVFILE)

tidy:
	go mod tidy -v

lint:
	go tool golangci-lint run --fix

build:
	go build -o ccplan .

ci: fmt fix lint build cov

# Go Fix (modernize)
fix:
	go fix ./...
