package log

import (
	"log"
	"os"
)

var logError = log.New(os.Stderr, "[ERROR]: ", log.LstdFlags)
var logInfo = log.New(os.Stdout, "[INFO]: ", log.LstdFlags)

func Errorln(v ...any) {
	logError.Println(v...)
}

func Errorf(format string, v ...any) {
	logError.Printf(format, v...)
}

func Infoln(v ...any) {
	logInfo.Println(v...)
}

func Infof(format string, v ...any) {
	logInfo.Printf(format, v...)
}

func Fatal(v ...any) {
	logInfo.Println(v...)
	os.Exit(1)
}

func Println(v ...any) {
	logInfo.Println(v...)
}

func Printf(format string, v ...any) {
	logInfo.Printf(format, v...)
}
