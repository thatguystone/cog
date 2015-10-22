package clog

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestMain(m *testing.M) {
	check.Main(m)
}

func basicTestConfig(c *check.C) Config {
	return Config{
		Outputs: map[string]*ConfigOutput{
			"test": {
				Which: "file",
				Level: Debug,
				Args: ConfigOutputArgs{
					"path": c.FS.Path("test"),
				},
			},
		},
		Modules: map[string]*ConfigModule{
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

	c.False(lg.EnabledFor(Debug))
	c.True(lg.EnabledFor(Info))

	lg.Debug("Debug")
	lg.Info("Info")
	lg.Warn("Warn")
	lg.Error("Error")

	c.Panic(func() {
		lg.Panic("Panic")
	})

	plain := c.FS.SReadFile("plain")
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
	cfg.Outputs["doesntExist"] = &ConfigOutput{
		Which: "doesntExist",
	}

	cfg.Outputs["errOut"] = &ConfigOutput{
		Which: "errOut",
	}

	cfg.Outputs["badFilters"] = &ConfigOutput{
		Which:   "file",
		Filters: []string{"iDontExist"},
		Args: ConfigOutputArgs{
			"path": c.FS.Path("badFilters"),
		},
	}

	err = l.Reconfigure(cfg)
	c.Error(err)

	cfg = basicTestConfig(c)
	cfg.Modules["test"] = cfg.Modules[""]
	cfg.Modules["test."] = cfg.Modules[""]

	cfg.Modules["noOuts"] = &ConfigModule{}
	cfg.Modules["badFilters"] = &ConfigModule{
		Outputs: []string{"file"},
		Filters: []string{"iDontExist"},
	}
	cfg.Modules["badOut"] = &ConfigModule{
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

	err = l.Reopen()
	c.MustNotError(err)

	lg.Info("after")

	c.Contains(c.FS.SReadFile("test.1"), "before")
	c.Contains(c.FS.SReadFile("test"), "after")
}

func TestDontPropagate(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)
	cfg.Modules["test"] = &ConfigModule{
		Outputs:       []string{"test"},
		Level:         Info,
		DontPropagate: true,
	}

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test")
	c.False(lg.EnabledFor(Debug))

	lg.Info("dont propagate")

	c.Equal(1,
		strings.Count(
			c.FS.SReadFile("test"),
			`dont propagate`))
}
