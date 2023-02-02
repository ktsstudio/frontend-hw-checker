FROM golang:1.16.5-buster as builder
RUN apt update -y && apt upgrade -y
RUN echo "deb http://http.us.debian.org/debian/ testing non-free contrib main" >> /etc/apt/sources.list
RUN apt-get install -y lsb-release > /dev/null 2>&1
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -

WORKDIR /code
COPY . .
RUN unset GOPATH && go build -o build/main .

FROM node:slim
# We don't need the standalone Chromium
ENV PUPPETEER_SKIP_CHROMIUM_DOWNLOAD true

# Install Google Chrome Stable and fonts
# Note: this installs the necessary libs to make the browser work with Puppeteer.
RUN apt-get update -y && apt upgrade -y && apt-get install gnupg wget -y && \
    wget --quiet --output-document=- https://dl-ssl.google.com/linux/linux_signing_key.pub | gpg --dearmor > /etc/apt/trusted.gpg.d/google-archive.gpg && \
    sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list' && \
    apt-get update && \
    apt-get install google-chrome-stable -y --no-install-recommends && \
    apt install -y git && \
    rm -rf /var/lib/apt/lists/*

ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/google-chrome-stable

WORKDIR /code
RUN node -v
RUN yarn -v
COPY --from=builder /code/build /code/build
