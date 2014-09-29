PID_FILE=/tmp/ustackd-test.pid
TEST_SERVER=./ustackd -c config/test.conf

build: prepare
	go build

prepare:
	go get -u code.google.com/p/gcfg \
		github.com/codegangsta/cli \
		github.com/mattn/go-sqlite3

vet:
	go get -u code.google.com/p/go.tools/cmd/vet
	go vet ./...

test: clean build vet
	# requires started server
	${TEST_SERVER} &
	sleep 1
	go test ./...
	ok=$$?
	sh -c "kill -INT `cat ${PID_FILE}`"
	exit ${ok}

fmt:
	go fmt ./...

cert:
	openssl req -x509 -newkey rsa:2048 -keyout config/key.pem -out config/cert.pem -days 365
	openssl rsa -in config/key.pem -out config/key.pem

clean:
	go clean
	rm -f ustackd.db ustackd.pid ${PID_FILE}

tserver: clean build
	${TEST_SERVER}