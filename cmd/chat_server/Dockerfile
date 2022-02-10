FROM golang:latest

ARG CHAT_LOG
ENV CHAT_LOG=${CHAT_LOG}

WORKDIR /app
COPY ./pkg ./pkg
COPY ./chat_server ./chat_server
COPY ./cmd/chat_server/main.go ./cmd/chat_server/main.go
COPY ./go.mod .
COPY ./go.sum .
#using old DB instead of re-creating it
# COPY ./db.sqlite ./db.sqlite

RUN go mod download
RUN go build ./cmd/chat_server/main.go

EXPOSE 8989

CMD ./main ${CHAT_LOG}}