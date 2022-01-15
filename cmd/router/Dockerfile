FROM golang:latest

WORKDIR /app
COPY ./pkg ./pkg
COPY ./router ./router
COPY ./cmd/router/main.go ./cmd/router/main.go
COPY ./go.mod .
COPY ./go.sum .

RUN go mod download
RUN go build ./cmd/router/main.go

EXPOSE 7777

CMD [ "./main" ]