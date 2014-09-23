build:
	go build

dependencies:
	go get -u code.google.com/p/gcfg

fmt:
	cd backends && go fmt
	cd config && go fmt
	cd connection && go fmt
	go fmt

test:
	cd backends && go test
	cd config && go test
	cd connection && go test
	go test
