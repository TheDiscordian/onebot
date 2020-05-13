// Copyright (c) 2020, The OneBot Contributors. All rights reserved.
package onelib

import (
	"io"
	"log"
	"os"
)

const (
	DEBUG = 1
)

type logger struct {
	*log.Logger
}

var (
	LogFile string
	Error   *logger
	Info    *logger
	Debug   *logger
)

func InitLoggers() {
	var (
		file *os.File
		err  error
	)

	file, err = os.OpenFile(LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
	}

	Error = &logger{log.New(io.MultiWriter(file, os.Stderr), "[error] ", log.Ldate|log.Ltime|log.Lshortfile)}
	Info = &logger{log.New(io.MultiWriter(file, os.Stdout), "[info]  ", log.Ldate|log.Ltime)}
	if DEBUG > 0 {
		Debug = &logger{log.New(os.Stdout, "[debug] ", log.Ltime)}
	}
}
