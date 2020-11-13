package logger

import (
	"os"
	"path/filepath"
	"testing"
)

func getLogFile() *os.File {
	logFile, _ := os.OpenFile(filepath.Join("/tmp", "loftsman-tests-logger.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	return logFile
}

func TestNew(t *testing.T) {
	New(getLogFile(), "loftsman-tests-logger")
}

func TestGetHelpLogo(t *testing.T) {
	GetHelpLogo()
}

func TestHeader(t *testing.T) {
	l := New(getLogFile(), "loftsman-tests-logger")
	l.Header("test")
}

func TestSubHeader(t *testing.T) {
	l := New(getLogFile(), "loftsman-tests-logger")
	l.SubHeader("test")
}

func TestClosingHeader(t *testing.T) {
	l := New(getLogFile(), "loftsman-tests-logger")
	l.ClosingHeader("test")
}
