package scaleway

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	sdkLogger "github.com/scaleway/scaleway-sdk-go/logger"
)

// logger is the implementation of the SDK Logger interface for this terraform plugin.
//
// cf. https://godoc.org/github.com/scaleway/scaleway-sdk-go/logger#Logger
type logger struct {
}

// l is the global logger singleton
var l = logger{}

// Debugf logs to the DEBUG log. Arguments are handled in the manner of fmt.Printf.
func (l logger) Debugf(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

// Infof logs to the INFO log. Arguments are handled in the manner of fmt.Printf.
func (l logger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

// Warningf logs to the WARNING log. Arguments are handled in the manner of fmt.Printf.
func (l logger) Warningf(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

// Errorf logs to the ERROR log. Arguments are handled in the manner of fmt.Printf.
func (l logger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

// ShouldLog allow the SDK to log only in DEBUG or TRACE levels.
func (l logger) ShouldLog(level sdkLogger.LogLevel) bool {
	return logging.IsDebugOrHigher()
}
