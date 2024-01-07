package tui

import (
	// "github.com/mattn/go-colorable"
   "os"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (t *TUI) setupLogger() *zap.Logger {
	zapEncoder := zap.NewDevelopmentEncoderConfig()
   zapEncoder.EncodeLevel = zapcore.CapitalColorLevelEncoder
   logFile, _ := os.OpenFile("./log/log.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
   syncer := zap.CombineWriteSyncers(logFile/*zapcore.AddSync(t.Logstream), /*os.Stdout/*, zapcore.AddSync(colorable.NewColorableStdout())*/)
   l := zap.New(zapcore.NewCore(
      zapcore.NewConsoleEncoder(zapEncoder),
      syncer,
      zapcore.DebugLevel,
   ), zap.AddCaller())
   l.Warn("cc")


   return l
}