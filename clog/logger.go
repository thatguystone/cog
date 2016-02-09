package clog

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tchap/go-patricia/patricia"
	"github.com/thatguystone/cog/stack"
)

// Logger provides a shared logger log; that is, any changes to this logger
// affect all other loggers of the same name. Changes only persist, however,
// while there is at least 1 reference to the logger logger.
type Logger struct {
	*logger
}

type goLogger struct {
	l   *Logger
	lvl Level
}

type logger struct {
	l    *Log
	pfx  patricia.Prefix
	key  string
	refs uint
	mod  *module
}

func newLogger(lg *logger) *Logger {
	lg.refs++
	pub := &Logger{lg}

	runtime.SetFinalizer(pub, finalizeLogger)

	return pub
}

func finalizeLogger(pub *Logger) {
	pub.l.mtx.Lock()
	defer pub.l.mtx.Unlock()

	pub.refs--
	if pub.refs == 0 {
		delete(pub.l.active, pub.key)
	}
}

// I feel dirty...
func (gl goLogger) Write(b []byte) (n int, err error) {
	n = len(b)
	gl.l.log(gl.lvl, stack.CallerAbove(1, "log"), string(b))
	return
}

// AsGoLogger gets a new *log.Logger that outputs to this logger at the given
// level. The `sub` parameter is used for creating a new sub-logger, as with
// Get().
func (lg *logger) AsGoLogger(sub string, lvl Level) *log.Logger {
	gl := goLogger{
		l:   lg.Get(sub),
		lvl: lvl,
	}

	return log.New(gl, "", 0)
}

func (lg *logger) LogEntry(e Entry) {
	e.Depth++
	e.Time = time.Now()
	e.Src = stack.Caller(e.Depth)

	e.Msg = strings.TrimSpace(e.Msg)

	e.Module = lg.key
	if e.Module == "" {
		e.Module = "(root)"
	}

	lg.l.rwmtx.RLock()
	mod := lg.mod
	lg.l.rwmtx.RUnlock()

	for mod != nil {
		if mod.filts.accept(e) {
			for _, o := range mod.outs {
				if !o.filts.accept(e) {
					continue
				}

				b, err := o.FormatEntry(e)
				if err == nil {
					err = o.Write(b)
				}

				if err != nil && !e.ignoreErrors {
					// This should never happen, so screw efficiency at this
					// point
					lg.l.Get("").LogEntry(Entry{
						Level: Error,
						Depth: 1,
						Msg: fmt.Sprintf("failed to write log entry to %s: %v",
							o.String(),
							err),
						ignoreErrors: true,
					})
				}
			}
		}

		if mod.dontPropagate {
			break
		}

		mod = mod.parent
	}

	if e.Level == Panic {
		panic(errors.New(e.Msg))
	}

	if e.Level == Fatal {
		os.Exit(2)
	}
}

// Assums a write lock on lg.l.rwmtx is held
func (lg *logger) updateModule(tr *patricia.Trie) {
	mod := tr.Get(lg.pfx)
	if mod != nil {
		lg.mod = mod.(*module)
	} else {
		tr.VisitPrefixes(lg.pfx,
			func(_ patricia.Prefix, item patricia.Item) error {
				lg.mod = item.(*module)
				return nil
			})
	}
}

func (lg *logger) log(lvl Level, depth int, args ...interface{}) {
	lg.LogEntry(Entry{
		Level: lvl,
		Depth: depth + 1,
		Msg:   strings.TrimSpace(fmt.Sprintln(args...)),
	})
}

func (lg *logger) logData(lvl Level, depth int, d Data, format string, args ...interface{}) {
	lg.LogEntry(Entry{
		Level: lvl,
		Depth: depth + 1,
		Data:  d,
		Msg:   fmt.Sprintf(format, args...),
	})
}

func (lg *logger) logFormat(lvl Level, depth int, format string, args ...interface{}) {
	lg.logData(lvl, depth+1, nil, format, args...)
}

func (lg *logger) Debug(args ...interface{}) {
	lg.log(Debug, 1, args...)
}

func (lg *logger) Debugd(d Data, format string, args ...interface{}) {
	lg.logData(Debug, 1, d, format, args...)
}

func (lg *logger) Debugf(format string, args ...interface{}) {
	lg.logFormat(Debug, 1, format, args...)
}

func (lg *logger) Info(args ...interface{}) {
	lg.log(Info, 1, args...)
}

func (lg *logger) Infod(d Data, format string, args ...interface{}) {
	lg.logData(Info, 1, d, format, args...)
}

func (lg *logger) Infof(format string, args ...interface{}) {
	lg.logFormat(Info, 1, format, args...)
}

func (lg *logger) Warn(args ...interface{}) {
	lg.log(Warn, 1, args...)
}

func (lg *logger) Warnd(d Data, format string, args ...interface{}) {
	lg.logData(Warn, 1, d, format, args...)
}

func (lg *logger) Warnf(format string, args ...interface{}) {
	lg.logFormat(Warn, 1, format, args...)
}

func (lg *logger) Error(args ...interface{}) {
	lg.log(Error, 1, args...)
}

func (lg *logger) Errord(d Data, format string, args ...interface{}) {
	lg.logData(Error, 1, d, format, args...)
}

func (lg *logger) Errorf(format string, args ...interface{}) {
	lg.logFormat(Error, 1, format, args...)
}

func (lg *logger) Panic(args ...interface{}) {
	lg.log(Panic, 1, args...)
}

func (lg *logger) Panicd(d Data, format string, args ...interface{}) {
	lg.logData(Panic, 1, d, format, args...)
}

func (lg *logger) Panicf(format string, args ...interface{}) {
	lg.logFormat(Panic, 1, format, args...)
}

func (lg *logger) Fatal(args ...interface{}) {
	lg.log(Fatal, 1, args...)
}

func (lg *logger) Fatald(d Data, format string, args ...interface{}) {
	lg.logData(Fatal, 1, d, format, args...)
}

func (lg *logger) Fatalf(format string, args ...interface{}) {
	lg.logFormat(Fatal, 1, format, args...)
}

// Get gets `sub` as a child of this logger. That is, if the current logger is
// something like "base", and this is called sub="child", it would return a
// logger for the module "base.child".
func (lg *logger) Get(sub string) *Logger {
	_, sub = modulePrefix(sub)
	return lg.l.Get(fmt.Sprintf("%s.%s", lg.key, sub))
}
