FROM golang:1.16.5-buster as builder
RUN apt update -y && apt upgrade -y
RUN echo "deb http://http.us.debian.org/debian/ testing non-free contrib main" >> /etc/apt/sources.list
RUN apt-get install -y lsb-release > /dev/null 2>&1
RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash -

WORKDIR /code
COPY . .
RUN unset GOPATH && go build -o build/main .

FROM node:18-slim
# We don't need the standalone Chromium
ENV PUPPETEER_SKIP_CHROMIUM_DOWNLOAD true

# Install dependencies
RUN apt-get update && \
    apt-get install -y wget gnupg git

# Download and install Google Chrome
ENV CHROME_VERSION=120.0.6099.129-1
RUN wget -q https://dl.google.com/linux/chrome/deb/pool/main/g/google-chrome-stable/google-chrome-stable_${CHROME_VERSION}_amd64.deb
RUN apt-get -y update
RUN apt-get install -y ./google-chrome-stable_${CHROME_VERSION}_amd64.deb

ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/google-chrome-stable

WORKDIR /code
RUN node -v
RUN yarn -v
COPY --from=builder /code/build /code/build
