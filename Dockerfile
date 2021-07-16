FROM golang:1.16.5-buster
WORKDIR code
COPY . .
RUN unset GOPATH && go build -o build/main *.go
RUN rm * || true