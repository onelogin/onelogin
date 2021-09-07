LATEST_GIT_TAG := $(shell git describe --abbrev=0 --tags)

test:
	go get github.com/jpoles1/gopherbadger
	gopherbadger -md="readme.md" -png=false

secure:
	curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s
	./bin/gosec -exclude=G104,G109 ./...

ship:
	go build ./... && go install .
	if [ "$(shell onelogin version)" != "${LATEST_GIT_TAG}" ]; then exit 1; fi
	bash ship.sh github.com/onelogin/onelogin

clear-tf:
	rm -rf .terraform/ && rm .terraform.lock.hcl terraform.* main.tf

install:
	go build ./... && go install .