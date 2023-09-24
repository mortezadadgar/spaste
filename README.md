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
- [X] Show a dialog upon copying to clipboard
- [X] Unit testing
- [ ] Allow custom paste address
- [ ] Use goose for sql migration
- [ ] Rewrite Dockerfile
- [ ] Integration tests
- [ ] Select or drop and drag text file
- [ ] Expire pastes after a week
