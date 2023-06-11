FROM alpine:latest
RUN apk --no-cache add go gcc
WORKDIR /usr/src/app
COPY . ./
RUN go build ./
EXPOSE 8080
CMD ["./spaste"]
