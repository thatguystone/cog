package clog

import (
	"runtime"
	"strings"
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestLogCoverage(t *testing.T) {
	c := check.New(t)
	cfg := basicTestConfig(c)

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test")

	lg.Debug("one", 2, "three")
	lg.WithKV("debug-data", 1).Debugf("Debug %s", "data")
	lg.Debugf("Debug#%s", "fun")
	lg.Info("one", 2, "three")
	lg.WithKV("info-data", 1).Infof("Info %s", "data")
	lg.Infof("Info#%s", "fun")
	lg.Warn("one", 2, "three")
	lg.WithKV("warn-data", 1).Warnf("Warn %s", "data")
	lg.Warnf("Warn#%s", "fun")
	lg.Error("one", 2, "three")
	lg.WithKV("error-data", 1).Errorf("Error %s", "data")
	lg.Errorf("Error#%s", "fun")

	c.Panics(func() {
		lg.Panic("one", 2, "three")
	})

	c.Panics(func() {
		lg.WithKV("panic-data", 1).Panicf("Panic %s", "data")
	})

	c.Panics(func() {
		lg.Panicf("Panic#%s", "fun")
	})

	test := c.FS.SReadFile("test")

	c.Equal(5, strings.Count(test, "one 2 three"))
	c.Equal(5, strings.Count(test, "data."))
	c.Equal(5, strings.Count(test, "#fun"))
}

func TestLogData(t *testing.T) {
	c := check.New(t)
	cfg := basicTestConfig(c)

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test")

	lg.
		With(Data{"a": 1, "b": 2, "c": 3}).
		WithKV("a", 4).
		With(Data{"b": 10}).
		Info("data!")

	test := c.FS.SReadFile("test")

	c.Contains(test, "data.a=4")
	c.Contains(test, "data.b=10")
	c.Contains(test, "data.c=3")
	c.Contains(test, "msg=data!")
}

func TestLogGet(t *testing.T) {
	c := check.New(t)
	cfg := basicTestConfig(c)

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get(" .base. ")
	lg.Info("first")

	lg = lg.Get(" .sub. ")
	lg.Info("second")

	runtime.GC()

	c.Equal(1,
		strings.Count(
			c.FS.SReadFile("test"),
			`level=info module=base msg=first`))
	c.Equal(1,
		strings.Count(
			c.FS.SReadFile("test"),
			`level=info module=base.sub msg=second`))
}

func TestLogAsGoLogger(t *testing.T) {
	c := check.New(t)
	cfg := basicTestConfig(c)

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test").AsGoLogger("sub", Info)
	lg.Println("MERP")
	lg.Println("HERP")

	test := c.FS.SReadFile("test")
	c.Log(test)
	c.Equal(2, strings.Count(test, "src=log_test.go"))
	c.Equal(2, strings.Count(test, "level=info"))
}
