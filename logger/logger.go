package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type logLevel int

const (
	lDebug logLevel = 0
	lInfo  logLevel = 1
	lWarn  logLevel = 2
	lError logLevel = 3
	lFatal logLevel = 4
)

type ctxKey int

const (
	TraceId ctxKey = iota
)

type logMessage struct {
	level   logLevel
	message string
	ctx     context.Context
	data    interface{}
}

func WithData(data interface{}) logMessage {
	return logMessage{data: data}
}

func (logMsg logMessage) WithData(data interface{}) logMessage {
	logMsg.data = data
	return logMsg
}

func Debug(ctx context.Context, message string) {
	logMsg := logMessage{level: lDebug, message: message, ctx: ctx}
	printLog(logMsg)
}

func (logMsg logMessage) Debug(ctx context.Context, message string) {
	logMsg.level = lDebug
	logMsg.message = message
	logMsg.ctx = ctx
	printLog(logMsg)
}

func Info(ctx context.Context, message string) {
	logMsg := logMessage{level: lInfo, message: message, ctx: ctx}
	printLog(logMsg)
}

func (logMsg logMessage) Info(ctx context.Context, message string) {
	logMsg.level = lInfo
	logMsg.message = message
	logMsg.ctx = ctx
	printLog(logMsg)
}

func Warn(ctx context.Context, message string) {
	logMsg := logMessage{level: lWarn, message: message, ctx: ctx}
	printLog(logMsg)
}

func (logMsg logMessage) Warn(ctx context.Context, message string) {
	logMsg.level = lWarn
	logMsg.message = message
	logMsg.ctx = ctx
	printLog(logMsg)
}

func Error(ctx context.Context, message string) {
	logMsg := logMessage{level: lError, message: message, ctx: ctx}
	printLog(logMsg)
}

func (logMsg logMessage) Error(ctx context.Context, message string) {
	logMsg.level = lError
	logMsg.message = message
	logMsg.ctx = ctx
	printLog(logMsg)
}

func Fatal(ctx context.Context, message string) {
	logMsg := logMessage{level: lFatal, message: message, ctx: ctx}
	printLog(logMsg)
}

func (logMsg logMessage) Fatal(ctx context.Context, message string) {
	logMsg.level = lFatal
	logMsg.message = message
	logMsg.ctx = ctx
	printLog(logMsg)
}

func printLog(logMsg logMessage) {
	builder := strings.Builder{}

	now := time.Now()
	builder.WriteString(now.Format("[2006-01-02T15:04:05.000-0700]"))

	switch logMsg.level {
	case lDebug:
		builder.WriteString("[DEBUG]")
	case lInfo:
		builder.WriteString("[INFO] ")
	case lWarn:
		builder.WriteString("[WARN] ")
	case lError:
		builder.WriteString("[ERROR]")
	case lFatal:
		builder.WriteString("[FATAL]")
	}

	traceId := ""

	if logMsg.ctx != nil {
		ctxTraceId := logMsg.ctx.Value(TraceId)
		if ctxTraceId != nil {
			traceId = ctxTraceId.(string)
		}
	}

	if traceId != "" {
		builder.WriteRune('[')
		builder.WriteString(traceId)
		builder.WriteRune(']')
	}

	builder.WriteRune(' ')
	builder.WriteString(logMsg.message)

	if logMsg.data != nil {
		builder.WriteRune('\n')
		builder.WriteString(fmt.Sprintf("%+v", logMsg.data))
	}

	switch logMsg.level {
	case lDebug, lInfo, lWarn:
		fmt.Fprintln(os.Stderr, builder.String())
	case lError, lFatal:
		fmt.Fprintln(os.Stderr, builder.String())
	}

	if logMsg.level == lFatal {
		os.Exit(1)
	}
}
