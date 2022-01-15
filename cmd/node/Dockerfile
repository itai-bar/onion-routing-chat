FROM golang:latest

WORKDIR /app
COPY ./pkg ./pkg
COPY ./node ./node
COPY ./cmd/node/main.go ./cmd/node/main.go
COPY ./go.mod .
COPY ./go.sum .

RUN go mod download
RUN go build ./cmd/node/main.go

EXPOSE 8989

CMD [ "./main" ]