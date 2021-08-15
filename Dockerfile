FROM golang:1.16.5-buster
RUN apt update -y && apt upgrade -y && apt install -y git
RUN echo "deb http://http.us.debian.org/debian/ testing non-free contrib main" >> /etc/apt/sources.list
RUN apt update -y && apt install -y python3 python3-pip
WORKDIR code
COPY . .
RUN unset GOPATH && go build -o build/main *.go
RUN rm * || true
