FROM golang
RUN go get github.com/lxfontes/scanner/cmd/webscanner
RUN mkdir /static
COPY ./static /static
WORKDIR /
ENTRYPOINT ["/go/bin/webscanner","-endpoint", "0.0.0.0:8080"]
EXPOSE 8080
