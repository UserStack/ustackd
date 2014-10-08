FROM  grengojbo/go:latest
MAINTAINER Jens Bissinger "mail@jens-bissinger."

ADD . /go/src/github.com/UserStack/ustackd
WORKDIR /go/src/github.com/UserStack/ustackd
RUN make prepare

CMD ["go", "run", "ustackd.go", "-f"]
