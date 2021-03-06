PID_FILE=/tmp/ustackd-test.pid

build: prepare
	go build

prepare:
	go get -u code.google.com/p/gcfg \
		github.com/codegangsta/cli \
		github.com/mattn/go-sqlite3 \
		github.com/lib/pq \
		github.com/go-sql-driver/mysql

vet:
	go get -u code.google.com/p/go.tools/cmd/vet
	go vet ./...

test: clean vet sqlite

benchmark: clean
	go test -bench . -run benchmark -benchmem

sqlite:
	go test -v ./...
	
ci: test postgresql mysql
	
# tailored to travis ci
postgresql:
	psql -c 'create database ustackd;' -U postgres
	TEST_CONFIG=config/test_psql.conf go test -v ./...
	psql -c 'drop database ustackd;' -U postgres

# tailored to travis ci
mysql:
	mysql -e 'create database ustackd;'
	TEST_CONFIG=config/test_mysql.conf go test -v ./...
	mysql -e 'drop database ustackd;'

fmt:
	go fmt ./...

cert:
	openssl req -x509 -newkey rsa:2048 -keyout config/key.pem -out config/cert.pem -days 365
	openssl rsa -in config/key.pem -out config/key.pem

clean:
	go clean
	rm -f ustackd.db ustackd.pid ${PID_FILE}

coverage:
	go get -u code.google.com/p/go.tools/cmd/cover
	go test ./... -cover
	go test ./server -coverprofile=c.out
	go tool cover -html=c.out -o=cover/server.html
	go test ./backends -coverprofile=c.out
	go tool cover -html=c.out -o=cover/backends.html
