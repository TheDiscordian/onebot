// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package onelib

import (
	"io"
	"log"
	"os"
)

const (
	// DEBUG is a relic constant, it will be removed in favour of something else in the future.
	DEBUG = 1
)

type logger struct {
	*log.Logger
}

var (
	// Error is used for logging errors. It outputs to stderr and file.
	Error *logger
	// Info is used for logging information. It outputs to stdout and file.
	Info *logger
	// Debug is used for logging miscellaneous things, mostly for debugging code. It outputs to stdout.
	Debug *logger
)

// InitLoggers is supposed to only be called once, it initializes the loggers, opening any related logfiles.
func InitLoggers(logfile string) {
	var (
		file *os.File
		err  error
	)

	file, err = os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
	}

	Error = &logger{log.New(io.MultiWriter(file, os.Stderr), "[error] ", log.Ldate|log.Ltime|log.Lshortfile)}
	Info = &logger{log.New(io.MultiWriter(file, os.Stdout), "[info]  ", log.Ldate|log.Ltime)}
	if DEBUG > 0 {
		Debug = &logger{log.New(os.Stdout, "[debug] ", log.Ltime)}
	}
}
