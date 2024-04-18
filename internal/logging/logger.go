package logging

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/scaleway/scaleway-sdk-go/logger"
)

func init() {
	logger.SetLogger(L)
}

// Logger is the implementation of the SDK Logger interface for this terraform plugin.
//
// cf. https://godoc.org/github.com/scaleway/scaleway-sdk-go/logger#Logger
type Logger struct{}

// L is the global Logger singleton
var L = Logger{}

// Debugf logs to the DEBUG log. Arguments are handled in the manner of fmt.Printf.
func (l Logger) Debugf(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

// Infof logs to the INFO log. Arguments are handled in the manner of fmt.Printf.
func (l Logger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

// Warningf logs to the WARNING log. Arguments are handled in the manner of fmt.Printf.
func (l Logger) Warningf(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

// Errorf logs to the ERROR log. Arguments are handled in the manner of fmt.Printf.
func (l Logger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

// Printf logs to the DEBUG log. Arguments are handled in the manner of fmt.Printf.
func (l Logger) Printf(format string, args ...interface{}) {
	l.Debugf(format, args...)
}

// ShouldLog allow the SDK to log only in DEBUG or TRACE levels.
func (l Logger) ShouldLog(_ logger.LogLevel) bool {
	return logging.IsDebugOrHigher()
}
