package tui

import (
	// "github.com/mattn/go-colorable"
	"os"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
	"fmt"
	"time"
)

// Setup a new logger
func (t *TUI) setupLogger() *zap.Logger {
	zapEncoder := zap.NewDevelopmentEncoderConfig()
	// Define colorful logger output
	zapEncoder.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// Define a log file to also write the logs too
	logFile, _ := os.OpenFile(t.Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// Define a string writer to redirect log output to StringWriter, which will copy it to logC channel which the TUI consumes to show log in text area
	uiLogWriter := StringWriter{ch: t.logC};
	// Add both file output and TUI output to logger's writers
	syncer := zap.CombineWriteSyncers(logFile, zapcore.AddSync(uiLogWriter)/*apcore.AddSync(colorable.NewColorableStdout())*/)
	l := zap.New(zapcore.NewCore(
    	zapcore.NewConsoleEncoder(zapEncoder),
      	syncer,
      	zapcore.DebugLevel,
   	), zap.AddCaller())

   return l
}

// StringWriter implements the io.Writer interface by writing to a channel
type StringWriter struct {
	ch chan<- string
}

func (sw StringWriter) Write(p []byte) (n int, err error) {
	str := string(p)
   	str, _ = simplifyLogFormat(str) // simplify log string for UI output
	sw.ch <- str // Write the string to the channel
	return len(p), nil
}

// log format simplifier for UI log, file recieves full log format
func simplifyLogFormat(input string) (string, error) {
	// Split the input string by tab ("\t") to separate its components
	parts := strings.Split(input, "\t")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid log format")
	}

	// Extract the timestamp from the first part of the input
	timestamp := parts[0]

	// Parse the timestamp to a time.Time object
	timeObj, err := time.Parse("2006-01-02T15:04:05.999-0700", timestamp)
	if err != nil {
		return "", err
	}

	// Format the timestamp to the desired time format
	timeFormatted := timeObj.Format("15:04:05")

	// Construct the output string in the desired format
	output := fmt.Sprintf(" [gray]%s[-] %s\t%s", timeFormatted, parts[1], parts[3])

	return output, nil
}