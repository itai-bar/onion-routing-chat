FROM golang:latest

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o /chatServer

EXPOSE 8989

CMD [ "/chatServer" ]