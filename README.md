# Simple Paste
dead simple pastebin written in Golang.

## Starting
there is a configured Dockerfile to start this project.
```
$ docker build -t spaste .
$ docker run -p 8080:8080 -t spaste
```

## TODO
- [ ] Password protected paste
- [ ] Show a dialog upon copying to clipboard
- [X] Unit testing
- [ ] Allow custom paste address
