FROM alpine:latest
RUN apk --no-cache add go
WORKDIR /usr/src/app
COPY . ./
RUN go build ./
RUN mkdir --parents /var/cache/db
EXPOSE 8080
CMD ["./spaste"]
