FROM golang:1.22.1

WORKDIR /app
COPY . .

RUN go mod tidy

RUN apt-get update && apt-get install -y gdal-bin