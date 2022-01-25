FROM golang:latest

ARG NODE_LOG
ENV NODE_LOG=${NODE_LOG}

WORKDIR /app
COPY ./pkg ./pkg
COPY ./node ./node
COPY ./cmd/node/main.go ./cmd/node/main.go
COPY ./go.mod .
COPY ./go.sum .

RUN go mod download
RUN go build ./cmd/node/main.go

EXPOSE 8989

CMD ./main ${NODE_LOG}