build:
	go build

prepare:
	go get -u code.google.com/p/gcfg \
		github.com/codegangsta/cli \
		github.com/mattn/go-sqlite3

fmt:
	cd backends && go fmt
	cd config && go fmt
	cd connection && go fmt
	cd server && go fmt
	go fmt

test:
	cd backends && go test
	cd config && go test
	cd connection && go test
	cd server && go test
	go test
