FROM docker.io/golang:latest

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod tidy && go mod verify

COPY . .
RUN go build -v -o server ./cmd/server/main.go

EXPOSE 80
EXPOSE 53

CMD ["/usr/src/app/server"]
