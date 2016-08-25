package clog

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/tchap/go-patricia/patricia"
	"github.com/iheartradio/cog/stack"
)

// Log provides a shared log.
type Log struct {
	*dLog
}

// Data log
type dLog struct {
	*mLog
	d Data
}

// Module log
type mLog struct {
	ctx   *Ctx
	pfx   patricia.Prefix
	key   string
	refs  uint
	mod   *module
	stats Stats
}

type goLogger struct {
	l   *Log
	lvl Level
}

func newLogger(l *dLog) *Log {
	l.refs++
	pub := &Log{l}

	runtime.SetFinalizer(pub, finalizeLogger)

	return pub
}

func finalizeLogger(pub *Log) {
	pub.ctx.mtx.Lock()
	defer pub.ctx.mtx.Unlock()

	pub.refs--
	if pub.refs == 0 {
		delete(pub.ctx.active, pub.key)
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
func (l *mLog) AsGoLogger(sub string, lvl Level) *log.Logger {
	gl := goLogger{
		l:   l.Get(sub),
		lvl: lvl,
	}

	return log.New(gl, "", 0)
}

func (l *mLog) LogEntry(e Entry) {
	e.Depth++
	e.Time = time.Now()
	e.Src = stack.Caller(e.Depth)

	e.Msg = strings.TrimSpace(e.Msg)

	e.Module = l.key
	if e.Module == "" {
		e.Module = "(root)"
	}

	if e.Host == "" {
		e.Host = Hostname()
	}

	l.ctx.rwmtx.RLock()
	mod := l.mod
	l.ctx.rwmtx.RUnlock()

	for mod != nil {
		if mod.filts.accept(e) {
			for _, o := range mod.outs {
				if !o.filts.accept(e) {
					continue
				}

				b, err := o.FormatEntry(e)
				if err != nil {
					if !e.ignoreErrors {
						o.logErr(fmt.Errorf("format failed: %v", err))
					}
					continue
				}

				o.Produce(b)
			}
		}

		if mod.dontPropagate {
			break
		}

		mod = mod.parent
	}

	atomic.AddInt64(&l.stats.Counts[e.Level], 1)

	if e.Level == Panic {
		panic(errors.New(e.Msg))
	}
}

// Assumes a write lock on l.ctx.rwmtx is held
func (l *mLog) updateModule(tr *patricia.Trie) {
	mod := tr.Get(l.pfx)
	if mod != nil {
		l.mod = mod.(*module)
	} else {
		tr.VisitPrefixes(l.pfx,
			func(_ patricia.Prefix, item patricia.Item) error {
				l.mod = item.(*module)
				return nil
			})
	}
}

// Get gets `sub` as a child of this logger. That is, if the current logger is
// something like "base", and this is called sub="child", it would return a
// logger for the module "base.child".
func (l *mLog) Get(sub string) *Log {
	_, sub = modulePrefix(sub)
	return l.ctx.Get(fmt.Sprintf("%s.%s", l.key, sub))
}

func (l *dLog) log(lvl Level, depth int, args ...interface{}) {
	l.LogEntry(Entry{
		Level: lvl,
		Depth: depth + 1,
		Data:  l.d,
		Msg:   strings.TrimSpace(fmt.Sprintln(args...)),
	})
}

func (l *dLog) format(lvl Level, depth int, format string, args ...interface{}) {
	l.LogEntry(Entry{
		Level: lvl,
		Depth: depth + 1,
		Data:  l.d,
		Msg:   strings.TrimSpace(fmt.Sprintf(format, args...)),
	})
}

func (l *dLog) With(d Data) Logger {
	dl := &dLog{
		mLog: l.mLog,
		d:    make(Data, len(d)+len(l.d)),
	}

	for k, v := range l.d {
		dl.d[k] = v
	}

	for k, v := range d {
		dl.d[k] = v
	}

	return dl
}

func (l *dLog) WithKV(k string, v interface{}) Logger {
	return l.With(Data{k: v})
}

func (l *dLog) Debug(args ...interface{}) {
	l.log(Debug, 1, args...)
}

func (l *dLog) Debugf(format string, args ...interface{}) {
	l.format(Debug, 1, format, args...)
}

func (l *dLog) Info(args ...interface{}) {
	l.log(Info, 1, args...)
}

func (l *dLog) Infof(format string, args ...interface{}) {
	l.format(Info, 1, format, args...)
}

func (l *dLog) Warn(args ...interface{}) {
	l.log(Warn, 1, args...)
}

func (l *dLog) Warnf(format string, args ...interface{}) {
	l.format(Warn, 1, format, args...)
}

func (l *dLog) Error(args ...interface{}) {
	l.log(Error, 1, args...)
}

func (l *dLog) Errorf(format string, args ...interface{}) {
	l.format(Error, 1, format, args...)
}

func (l *dLog) Panic(args ...interface{}) {
	l.log(Panic, 1, args...)
}

func (l *dLog) Panicf(format string, args ...interface{}) {
	l.format(Panic, 1, format, args...)
}
