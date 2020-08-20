package logger

import "github.com/sirupsen/logrus"

var log *logrus.Logger

// Options describe the initialization options for the shared logger singleton
type Options struct {
	Formatter *logrus.TextFormatter
	Level     logrus.Level
	Hooks     []logrus.Hook
}

// Initialize a new logger singleton. If one has already been instantiated, the singleton is returned
func Initialize(o *Options) *logrus.Logger {
	if log != nil {
		return log
	}

	log = logrus.New()

	if o.Formatter != nil {
		log.SetFormatter(o.Formatter)
	}

	if o.Level != 0 {
		log.SetLevel(o.Level)
	}

	if o.Hooks != nil {
		for _, h := range o.Hooks {
			log.AddHook(h)
		}
	}

	return log
}

// Log returns the logger singleton. Panics if the singleton has not been initialized
func Log() *logrus.Logger {
	if log == nil {
		panic("Logger cannot be used until it has been initialized!")
	}
	return log
}

// Reset the singleton so that Initialize must be called again
func Reset() {
	log = nil
}
