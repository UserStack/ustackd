PID_FILE=./ustackd.pid

build: prepare
	go build

prepare:
	go get -u code.google.com/p/gcfg \
		github.com/codegangsta/cli \
		github.com/mattn/go-sqlite3

test:
	# requires started server
	go run ustackd.go -f &
	sleep 1
	go test ./...
	ok=$$?
	sh -c "kill -INT `cat ${PID_FILE}`"
	exit ${ok}

fmt:
	go fmt ./...
