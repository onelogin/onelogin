test:
	go get github.com/jpoles1/gopherbadger
	gopherbadger -md="readme.md" -png=false

secure:
	curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s
	./bin/gosec -exclude=G104,G109 ./...

ship:
	bash ship.sh github.com/onelogin/onelogin

clear-tf:
	rm -rf .terraform/ && rm .terraform.lock.hcl terraform.* main.tf
