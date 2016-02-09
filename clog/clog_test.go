package clog

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/thatguystone/cog/check"
)

func TestMain(m *testing.M) {
	check.Main(m)
}

func basicTestConfig(c *check.C) Config {
	return Config{
		Outputs: map[string]*OutputConfig{
			"test": {
				Which: "file",
				Level: Debug,
				Args: ConfigArgs{
					"path": c.FS.Path("test"),
				},
			},
		},
		Modules: map[string]*ModuleConfig{
			"": {
				Outputs: []string{"test"},
				Level:   Debug,
			},
		},
	}
}

func TestBasic(t *testing.T) {
	c := check.New(t)

	j := fmt.Sprintf(
		`{
			"outputs": {
				"plain": {
					"which": "file",
					"level": "info",
					"args": {
						"path": %q
					}
				},
				"json": {
					"which": "jsonfile",
					"level": "info",
					"args": {
						"path": %q
					}
				},
				"plain-error": {
					"which": "file",
					"level": "error",
					"args": {
						"path": %q
					}
				}
			},
			"modules": {
				"": {
					"outputs": ["plain", "plain-error", "json"],
					"level": "debug"
				},
				"test": {
					"outputs": ["plain", "json"],
					"level": "debug"
				},
				"test.sub": {
					"outputs": ["plain"],
					"level": "debug"
				}
			}
		}`,
		c.FS.Path("plain"),
		c.FS.Path("json"),
		c.FS.Path("error"))

	l, err := NewFromJSONReader(strings.NewReader(j))
	c.MustNotError(err)

	lg := l.Get("test.sub")

	lg2 := l.Get("test.sub")
	c.Equal(lg, lg2)

	lg.Debug("Debug")
	lg.Info("Info")
	lg.Warn("Warn")
	lg.Error("Error")

	c.Panic(func() {
		lg.Panic("Panic")
	})

	plain := c.FS.SReadFile("plain")
	c.Equal(0,
		strings.Count(plain,
			`level=debug`))
	c.Equal(3,
		strings.Count(plain,
			`level=info module=test.sub msg=Info`))

	json := c.FS.SReadFile("json")
	c.Equal(2,
		strings.Count(json,
			`"level":"info"`))

	error := c.FS.SReadFile("error")
	c.Equal(0,
		strings.Count(error,
			`level=info`))
	c.Equal(1,
		strings.Count(error,
			`level=error`))

	j = fmt.Sprintf(
		`{
			"outputs": {
				"plain": {
					"which": "file",
					"level": "error",
					"args": {
						"path": %q
					}
				}
			},
			"modules": {
				"test.sub": {
					"outputs": ["plain"],
					"level": "debug"
				}
			}
		}`,
		c.FS.Path("error"))

	l.ReconfigureFromJSONReader(strings.NewReader(j))

	lg.Debug("AfterReconfig-Debug")
	lg.Info("AfterReconfig-Info")
	lg.Warn("AfterReconfig-Warn")
	lg.Error("AfterReconfig-Error")

	error = c.FS.SReadFile("error")
	c.Equal(0, strings.Count(error, `msg=AfterReconfig-Debug`))
	c.Equal(0, strings.Count(error, `msg=AfterReconfig-Info`))
	c.Equal(0, strings.Count(error, `msg=AfterReconfig-Warn`))
	c.Equal(1, strings.Count(error, `msg=AfterReconfig-Error`))

	c.Equal(0,
		strings.Count(
			c.FS.SReadFile("plain"),
			`AfterReconfig-Error`))
	c.Equal(0,
		strings.Count(
			c.FS.SReadFile("json"),
			`AfterReconfig-Error`))
}

func TestNewFromFileJSON(t *testing.T) {
	c := check.New(t)

	c.FS.SWriteFile("config.json",
		fmt.Sprintf(`{"file": %q}`, c.FS.Path("log")))

	l, err := NewFromFile(c.FS.Path("config.json"))
	c.MustNotError(err)

	l.ReconfigureFromFile(c.FS.Path("config.json"))
	c.MustNotError(err)
}

func TestNewFromFileErrors(t *testing.T) {
	c := check.New(t)

	c.FS.SWriteFile("config.merp", "merp")

	_, err := NewFromFile(c.FS.Path("config.merp"))
	c.Error(err)

	cfg := basicTestConfig(c)
	l, err := New(cfg)
	c.MustNotError(err)

	err = l.ReconfigureFromFile(c.FS.Path("config.merp"))
	c.Error(err)
}

func TestNoRoot(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)
	cfg.Modules["test"] = cfg.Modules[""]
	delete(cfg.Modules, "")

	l, err := New(cfg)
	c.MustNotError(err)

	l.Get("").Info("--root--")
	l.Get("test").Info("--not root--")

	test := c.FS.SReadFile("test")
	c.NotContains(test, "--root--")
	c.Contains(test, "--not root--")
}

func TestReconfigureErrors(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)
	l, err := New(cfg)
	c.MustNotError(err)

	cfg = basicTestConfig(c)
	cfg.Outputs["doesntExist"] = &OutputConfig{
		Which: "doesntExist",
	}

	cfg.Outputs["errOut"] = &OutputConfig{
		Which: "errOut",
	}

	cfg.Outputs["badFilters"] = &OutputConfig{
		Which: "file",
		Filters: []FilterConfig{
			FilterConfig{Which: "iDontExist"},
		},
		Args: ConfigArgs{
			"path": c.FS.Path("badFilters"),
		},
	}

	err = l.Reconfigure(cfg)
	c.Error(err)

	cfg = basicTestConfig(c)
	cfg.Modules["test"] = cfg.Modules[""]
	cfg.Modules["test."] = cfg.Modules[""]

	cfg.Modules["noOuts"] = &ModuleConfig{}
	cfg.Modules["badFilters"] = &ModuleConfig{
		Outputs: []string{"file"},
		Filters: []FilterConfig{
			FilterConfig{Which: "iDontExist"},
		},
	}
	cfg.Modules["badOut"] = &ModuleConfig{
		Outputs: []string{"rawr"},
	}

	err = l.Reconfigure(cfg)
	c.Error(err)
}

func TestReopen(t *testing.T) {
	c := check.New(t)

	l, err := New(basicTestConfig(c))
	c.MustNotError(err)

	lg := l.Get("test")

	lg.Info("before")

	os.Rename(c.FS.Path("test"), c.FS.Path("test.1"))

	err = l.Rotate()
	c.MustNotError(err)

	lg.Info("after")

	c.Contains(c.FS.SReadFile("test.1"), "before")
	c.Contains(c.FS.SReadFile("test"), "after")
}

func TestExit(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)

	cfg.File = c.FS.Path("default_file")
	out := cfg.Outputs["test"]
	out.Which = "exitOut"

	cfg.Modules["test"] = &ModuleConfig{
		Outputs: []string{"test", defaultConfigFileOutputName},
	}

	l, err := New(cfg)
	c.MustNotError(err)

	eo := l.outputs["test"].Outputter.(*exitOutput)
	l.Exit()
	c.True(eo.exited)

	err = l.Reconfigure(cfg)
	c.MustNotError(err)

	lg := l.Get("test")
	go func() {
		time.Sleep(time.Millisecond * 100)
		lg.Info("Delayed message")
	}()

	eo = l.outputs["test"].Outputter.(*exitOutput)
	l.Exit()
	c.True(eo.exited)
}

func TestDontPropagate(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)
	cfg.Modules["test"] = &ModuleConfig{
		Outputs:       []string{"test"},
		Level:         Info,
		DontPropagate: true,
	}

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test")
	lg.Debug("debug")
	lg.Info("dont propagate")

	c.Equal(0,
		strings.Count(
			c.FS.SReadFile("test"),
			`debug`))
	c.Equal(1,
		strings.Count(
			c.FS.SReadFile("test"),
			`dont propagate`))
}

func TestExtraSpaces(t *testing.T) {
	c := check.New(t)
	cfg := basicTestConfig(c)

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get(" module.spaces ")
	lg.Info("         tons of spaces           ")

	c.Log(c.FS.SReadFile("test"))

	c.Equal(1,
		strings.Count(
			c.FS.SReadFile("test"),
			`level=info module=module.spaces msg="tons of spaces"`))
}

func TestEmptyConfig(t *testing.T) {
	c := check.New(t)
	cfg := Config{}

	_, err := New(cfg)
	c.NotError(err)
}

func TestDefaultFile(t *testing.T) {
	c := check.New(t)
	cfg := Config{
		File: c.FS.Path("log"),
	}

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test")
	lg.Debugf("some %s", "log")
	lg.Infof("some %s", "log")

	c.Equal(1,
		strings.Count(
			c.FS.SReadFile("log"),
			`some log`))
}

func TestDefaultFileWithOthers(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)
	cfg.File = c.FS.Path("default_file")

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test")
	lg.Debugf("some %s", "log")
	lg.Infof("some %s", "log")

	c.Equal(1,
		strings.Count(
			c.FS.SReadFile("default_file"),
			`some log`))
	c.Equal(2,
		strings.Count(
			c.FS.SReadFile("test"),
			`some log`))
}

func TestDefaultTermLogger(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)
	cfg.Modules = nil

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test")
	lg.Debugf("some %s", "log")
	lg.Infof("some %s", "log")

	// Not really a way to test that this output right since it's going to the
	// console
}
