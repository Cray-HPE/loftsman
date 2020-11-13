// Package logger is a general-use logger utility for loftsman
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	colorTextReset = "\033[0m"
	colorCyan      = "\033[36m"
	colorWhite     = "\033[37m"
	colorGray      = "\033[37m"
	textBold       = "\033[1m"
	textUnderline  = "\033[4m"
)

var (
	record *strings.Builder
)

// Logger is our main logger object
type Logger struct {
	zerolog.Logger
}

type consoleWriter struct {
	zerologConsoleWriter zerolog.ConsoleWriter
}

func (cw consoleWriter) Write(p []byte) (int, error) {
	var line map[string]interface{}
	err := json.Unmarshal(p, &line)
	if err == nil {
		if _, ok := line["header"]; ok {
			return len(p), nil
		}
		if _, ok := line["sub-header"]; ok {
			return len(p), nil
		}
		if _, ok := line["closing-header"]; ok {
			return len(p), nil
		}
	}
	return cw.zerologConsoleWriter.Write(p)
}

// GetHelpLogo will return the logo to display in the CLI help
func GetHelpLogo() string {
	return fmt.Sprintf(`%s%s   _        __ _                                %s%s|\
  %s%s| | ___  / _| |_ ___ _ __ ___   __ _ _ __     %s%s| \
  %s%s| |/ _ \| |_| __/ __|  _   _ \ / _  |  _ \    %s%s|  \
  %s%s| | |_| |  _| |_\__ \ | | | | | |_| | | | |   %s%s|___\
  %s%s|_|\___/|_|  \__|___/_| |_| |_|\__,_|_| |_|  %s%s\--||___/
                                  %s~~~~~~~~~~~~~~%s\_____/%s~~~~~~~~~~%s`,
		textBold, colorWhite, colorTextReset, colorGray,
		textBold, colorWhite, colorTextReset, colorGray,
		textBold, colorWhite, colorTextReset, colorGray,
		textBold, colorWhite, colorTextReset, colorGray,
		textBold, colorWhite, colorTextReset, colorGray,
		colorCyan, colorGray, colorCyan, colorTextReset)
}

// Header will log a header for a section of other log messages
func (log *Logger) Header(text string) {
	fmt.Println(fmt.Sprintf(`%s         |\
         | \
         |  \
         |___\      %s%s%s%s%s
       \--||___/
  %s~~~~~~%s\_____/%s~~~~~~~
  %s`, colorGray, textBold, colorWhite, text, colorTextReset, colorGray, colorCyan, colorGray, colorCyan, colorTextReset))
	log.Log().Str("header", text).Msg("")
}

// SubHeader will print a subheader section in log output
func (log *Logger) SubHeader(text string) {
	fmt.Println("")
	fmt.Println(fmt.Sprintf("%s%s%s", colorCyan, strings.Repeat("~", len(text)+5), colorTextReset))
	fmt.Println(fmt.Sprintf("%s%s%s", textBold, text, colorTextReset))
	fmt.Println(fmt.Sprintf("%s%s%s", colorCyan, strings.Repeat("~", len(text)+5), colorTextReset))
	fmt.Println("")
	log.Log().Str("sub-header", text).Msg("")
}

// ClosingHeader will precede closing/final/called-out log lines
func (log *Logger) ClosingHeader(text string) {
	fmt.Println("")
	fmt.Println(fmt.Sprintf("%s%s%s%s", textBold, textUnderline, text, colorTextReset))
	fmt.Println("")
	log.Log().Str("closing-header", text).Msg("")
}

// GetRecord will return the saved log record from any single cli run
func (log *Logger) GetRecord() string {
	return record.String()
}

// New will return a new instance of a logger
func New(jsonLogFile *os.File, commandName string) *Logger {
	record = &strings.Builder{}
	zerologConsoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	multiWriter := io.MultiWriter(consoleWriter{zerologConsoleWriter: zerologConsoleWriter}, jsonLogFile, record)
	return &Logger{
		zerolog.New(multiWriter).With().Str("command", commandName).Timestamp().Logger(),
	}
}
