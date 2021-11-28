FROM golang:latest

WORKDIR /app
COPY . .
COPY ../pkg .

RUN go mod download
RUN go build -o /node

EXPOSE 8989

CMD [ "/node" ]