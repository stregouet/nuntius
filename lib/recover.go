package lib

import (
    "fmt"
    "os"
	"runtime/debug"
)



func Recover(logger *Logger, cb func()) {
	var err interface{}
	logger.Debug("recoverterm called")
	if err = recover(); err == nil {
		logger.Debug("recoverterm no error")
		return
	}
	logger.Debugf("recoverterm error %v", err)
	debug.PrintStack()
    if cb != nil {
        cb()
    }
	fmt.Fprintf(os.Stderr, "nuntius crashed: %v\n", err)
	os.Exit(1)
}
