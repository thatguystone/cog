package clog

import (
	"runtime"
	"strings"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestLoggerCoverage(t *testing.T) {
	c := check.New(t)
	cfg := basicTestConfig(c)

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test")

	lg.Debug("one", 2, "three")
	lg.Debugd(Data{"debug-data": 1}, "Debug %s", "data")
	lg.Debugf("Debug#%s", "fun")
	lg.Info("one", 2, "three")
	lg.Infod(Data{"info-data": 1}, "Info %s", "data")
	lg.Infof("Info#%s", "fun")
	lg.Warn("one", 2, "three")
	lg.Warnd(Data{"warn-data": 1}, "Warn %s", "data")
	lg.Warnf("Warn#%s", "fun")
	lg.Error("one", 2, "three")
	lg.Errord(Data{"error-data": 1}, "Error %s", "data")
	lg.Errorf("Error#%s", "fun")

	c.Panic(func() {
		lg.Panic("one", 2, "three")
	})

	c.Panic(func() {
		lg.Panicd(Data{"panic-data": 1}, "Panic %s", "data")
	})

	c.Panic(func() {
		lg.Panicf("Panic#%s", "fun")
	})

	test := c.FS.SReadFile("test")

	c.Equal(5, strings.Count(test, "one 2 three"))
	c.Equal(5, strings.Count(test, "data."))
	c.Equal(5, strings.Count(test, "#fun"))
}

func TestLoggerGet(t *testing.T) {
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

func TestLoggerAsGoLogger(t *testing.T) {
	c := check.New(t)
	cfg := basicTestConfig(c)

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test").AsGoLogger("sub", Info)
	lg.Println("MERP")
	lg.Println("HERP")

	test := c.FS.SReadFile("test")
	c.Log(test)
	c.Equal(2, strings.Count(test, "src=logger_test.go"))
	c.Equal(2, strings.Count(test, "level=info"))
}
