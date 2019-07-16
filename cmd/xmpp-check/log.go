package main

import (
	"os"

	"github.com/bdlm/log"
	stdLogger "github.com/bdlm/std/logger"
)

type hook struct{}

func (h *hook) Fire(entry *log.Entry) error {
	switch entry.Level {
	case log.PanicLevel:
		entry.Logger.Out = os.Stderr
	case log.FatalLevel:
		entry.Logger.Out = os.Stderr
	case log.ErrorLevel:
		entry.Logger.Out = os.Stderr
	case log.WarnLevel:
		entry.Logger.Out = os.Stdout
	case log.InfoLevel:
		entry.Logger.Out = os.Stdout
	case log.DebugLevel:
		entry.Logger.Out = os.Stdout
	default:
	}

	return nil
}

func (h *hook) Levels() []stdLogger.Level {
	return log.AllLevels
}
