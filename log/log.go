package log

import (
	"log"
)

// Info log info
func Info(v ...interface{}) {
	log.Println("[info]:", v)
}

// Error log info
func Error(v ...interface{}) {
	log.Println("[Error]:", v)
}

// Warn log info
func Warn(v ...interface{}) {
	log.Println("[Warn]:", v)
}

func Fatal(v ...interface{}) {
	log.Fatal("[Error]:", v)
}
