FROM golang:latest

WORKDIR /app
COPY . .

RUN go mod download
RUN go build ./cmd/node/main.go 

EXPOSE 8989

CMD [ "./main" ]