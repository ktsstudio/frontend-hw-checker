FROM golang:1.16.5-buster
RUN apt update -y && apt upgrade -y && apt install -y git
RUN echo "deb http://http.us.debian.org/debian/ testing non-free contrib main" >> /etc/apt/sources.list
RUN apt-get install -y lsb-release > /dev/null 2>&1
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -
RUN apt-get install -y nodejs
RUN node -v
RUN npm install --global yarn
RUN yarn -v
WORKDIR code
COPY . .
RUN unset GOPATH && go build -o build/main .
RUN rm * || true
