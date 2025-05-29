FROM golang:alpine

RUN apk add --no-cache iputils curl

COPY . .

RUN go mod download
RUN go build -o /deviceapi 

EXPOSE 8080


CMD [ "/deviceapi" ]