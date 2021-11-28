FROM golang:latest

WORKDIR /app
COPY . .

RUN go mod download
RUN go build ./cmd/chat_server/main.go

EXPOSE 8989

CMD [ "./main" ]