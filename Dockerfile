FROM golang:1.16.5-buster
RUN apt update -y && apt upgrade -y && apt install -y python3 python3-pip git
WORKDIR code
COPY . .
RUN unset GOPATH && go build -o build/main *.go
RUN rm * || true
