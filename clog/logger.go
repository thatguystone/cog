package clog

import (
	"errors"
	"fmt"
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

type logger struct {
	l    *CLog
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

// EnabledFor checks if any Outputter in the chain might accept this message.
// Because some messages are filtered, it's impossible to tell with 100%
// accuracy if the message will be accepted until the full message is known.
func (lg *logger) EnabledFor(lvl Level) bool {
	lg.l.rwmtx.RLock()
	mod := lg.mod
	lg.l.rwmtx.RUnlock()

	for mod != nil {
		if mod.filts.levelEnabled(lvl) {
			for _, o := range mod.outs {
				if o.filts.levelEnabled(lvl) {
					return true
				}
			}
		}

		if mod.dontPropagate {
			break
		}

		mod = mod.parent
	}

	return false
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
