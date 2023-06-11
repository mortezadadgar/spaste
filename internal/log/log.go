package log

import (
	"log"
	"os"
)

var logError = log.New(os.Stderr, "[ERROR]: ", log.LstdFlags)
var logInfo = log.New(os.Stdout, "[INFO]: ", log.LstdFlags)

// Errorln calls log.Println with a prefix.
func Errorln(v ...any) {
	logError.Println(v...)
}

// Errorf calls log.Printf with a prefix.
func Errorf(format string, v ...any) {
	logError.Printf(format, v...)
}

// Infoln calls log.Println with a prefix.
func Infoln(v ...any) {
	logInfo.Println(v...)
}

// Infof calls log.Printf with a prefix.
func Infof(format string, v ...any) {
	logInfo.Printf(format, v...)
}

// Fatal calls log.Fatal with a prefix.
func Fatal(v ...any) {
	logError.Fatal(v...)
}

// Println calls log.Println with a prefix.
func Println(v ...any) {
	logInfo.Println(v...)
}

// Printf calls log.Printf with a prefix.
func Printf(format string, v ...any) {
	logInfo.Printf(format, v...)
}
