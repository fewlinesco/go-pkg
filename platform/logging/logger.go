package logging

import (
	"log"
	"os"
)

func NewDefaultLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
}
