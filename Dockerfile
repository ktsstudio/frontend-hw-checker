FROM golang:1.22

# Обновление пакетов и установка git
RUN apt update -y && apt upgrade -y && apt install -y git

# Добавление источников для debian testing
RUN echo "deb http://http.us.debian.org/debian/ testing non-free contrib main" >> /etc/apt/sources.list
RUN apt-get install -y lsb-release > /dev/null 2>&1

# Установка Node.js и Yarn
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
RUN apt-get install -y nodejs
RUN node -v
RUN npm install --global yarn
RUN yarn -v

# Установка зависимостей для Puppeteer и Chromium
RUN apt-get update && \
    apt-get install -y \
    wget \
    gnupg \
    libxss1 \
    libatk-bridge2.0-0 \
    libgtk-3-0 \
    libnss3 \
    libx11-xcb1 \
    chromium

# Установка рабочего каталога и сборка проекта
WORKDIR code
COPY . .
RUN unset GOPATH && go build -o build/main .
RUN rm * || true
