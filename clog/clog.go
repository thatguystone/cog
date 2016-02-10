package clog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/tchap/go-patricia/patricia"
	"github.com/thatguystone/cog"
	"github.com/thatguystone/cog/config"
)

// Log provides access to the logging facilities
type Log struct {
	rwmtx sync.RWMutex
	logState

	mtx    sync.Mutex
	active map[string]*logger
}

type logState struct {
	wg      *sync.WaitGroup
	outputs map[string]*output
	modules *patricia.Trie // Items of type *module
}

type module struct {
	pfx           patricia.Prefix
	parent        *module
	outs          []*output
	filts         filterSlice
	dontPropagate bool
}

type output struct {
	Outputter
	filts filterSlice
}

// New creates a new Log
func New(cfg Config) (l *Log, err error) {
	l = &Log{
		active: map[string]*logger{},
	}

	err = l.Reconfigure(cfg)
	if err != nil {
		l = nil
	}

	return
}

// NewFromFile creates a new Log, configured from the given file. The file type
// is determined by the extension.
func NewFromFile(path string) (l *Log, err error) {
	f, err := os.Open(path)

	if err == nil {
		defer f.Close()

		ext := filepath.Ext(strings.ToLower(path))
		switch ext {
		case ".json":
			l, err = NewFromJSONReader(f)

		default:
			err = fmt.Errorf("unsupported config file type: %s", ext)
		}
	}

	return
}

// NewFromJSONReader creates a new Log configured from JSON in the given
// reader.
func NewFromJSONReader(r io.Reader) (l *Log, err error) {
	d := json.NewDecoder(r)

	cfg := Config{}
	err = d.Decode(&cfg)

	if err == nil {
		l, err = New(cfg)
	}

	return
}

// Reconfigure reconfigures the entire logging system from the ground up. All
// active loggers are affected immediately, and all changes are applied
// atomically. If reconfiguration fails, the previous configuration remains.
func (l *Log) Reconfigure(cfg Config) error {
	wg := &sync.WaitGroup{}
	tmp := logState{
		wg:      wg,
		outputs: map[string]*output{},
		modules: patricia.NewTrie(),
	}

	if cfg.Outputs == nil {
		cfg.Outputs = map[string]*OutputConfig{}
	}

	if cfg.Modules == nil {
		cfg.Modules = map[string]*ModuleConfig{}
	}

	if cfg.File != "" {
		cfg.Outputs[defaultConfigFileOutputName] = &OutputConfig{
			Which: "jsonfile",
			Level: Info,
			Args: config.Args{
				"path": cfg.File,
			},
		}

		m, ok := cfg.Modules[""]
		if !ok {
			m = &ModuleConfig{
				Level: Info,
			}
			cfg.Modules[""] = m
		}

		m.Outputs = append(m.Outputs, defaultConfigFileOutputName)
	}

	if len(cfg.Modules) == 0 {
		cfg.Outputs[defaultTermOutputName] = &OutputConfig{
			Which: "term",
			Level: Info,
		}

		cfg.Modules[""] = &ModuleConfig{
			Outputs: []string{defaultTermOutputName},
			Level:   Info,
		}
	}

	es := cog.Errors{}
	for name, ocfg := range cfg.Outputs {
		o, err := newOutput(ocfg)
		if err != nil {
			es.Addf(err, `while creating output "%s"`, name)
			continue
		}

		filts, err := newFilters(ocfg.Level, ocfg.Filters)
		if err != nil {
			es.Addf(err, `for output "%s", invalid filter config`, name)
			continue
		}

		out := &output{
			Outputter: o,
			filts:     filts,
		}

		tmp.outputs[name] = out

		wg.Add(1)
		runtime.SetFinalizer(out, func(out *output) {
			go func() {
				out.Exit()
				for _, f := range out.filts {
					f.Exit()
				}

				wg.Done()
			}()
		})
	}

	if es.Empty() {
		for name, mcfg := range cfg.Modules {
			pfx, name := modulePrefix(name)

			if tmp.modules.Match(pfx) {
				es.Add(fmt.Errorf(`module "%s" already configured`, name))
				continue
			}

			if len(mcfg.Outputs) == 0 {
				es.Add(fmt.Errorf(`module "%s" has no outputs`, name))
				continue
			}

			filts, err := newFilters(mcfg.Level, mcfg.Filters)
			if err != nil {
				es.Addf(err, `for module "%s", invalid filter config`, name)
				continue
			}

			mod := &module{
				pfx:           pfx,
				filts:         filts,
				dontPropagate: mcfg.DontPropagate,
			}

			tmp.modules.Insert(pfx, mod)

			for _, oname := range mcfg.Outputs {
				o := tmp.outputs[oname]

				if o == nil {
					es.Add(fmt.Errorf(`unrecognized output "%s" for module "%s"`,
						oname,
						name))
				} else {
					mod.outs = append(mod.outs, o)
				}
			}
		}
	}

	err := es.Error()
	if err == nil {
		l.rwmtx.Lock()
		defer l.rwmtx.Unlock()

		tmp.modules.Visit(func(_ patricia.Prefix, item patricia.Item) error {
			var parent *module

			tmp.modules.VisitPrefixes(item.(*module).pfx,
				func(_ patricia.Prefix, item patricia.Item) error {
					mod := item.(*module)

					mod.parent = parent
					parent = mod

					return nil
				})

			return nil
		})

		for _, lg := range l.active {
			lg.updateModule(tmp.modules)
		}

		l.logState = tmp
	}

	return err
}

// ReconfigureFromJSONReader reconfigures the Log from the JSON in the given
// reader.
func (l *Log) ReconfigureFromJSONReader(r io.Reader) error {
	d := json.NewDecoder(r)

	cfg := Config{}
	err := d.Decode(&cfg)

	if err == nil {
		err = l.Reconfigure(cfg)
	}

	return err
}

// ReconfigureFromFile reconfigures the Log from the given file. The file type
// is determined by the extension.
func (l *Log) ReconfigureFromFile(path string) error {
	f, err := os.Open(path)

	if err == nil {
		defer f.Close()

		ext := filepath.Ext(strings.ToLower(path))
		switch ext {
		case ".json":
			err = l.ReconfigureFromJSONReader(f)

		default:
			err = fmt.Errorf("unsupported config file type: %s", ext)
		}
	}

	return err
}

// Exit can be used by servers that exit gracefully. It causes all Outputters
// to Exit and cleanup after themselves, and it blocks until done.
//
// This is typically the last thing you want to call before exiting. Also,
// calling Exit() is completely optional.
//
// This Log can be reused by calling Reconfigure().
func (l *Log) Exit() {
	l.rwmtx.RLock()
	wg := l.wg
	l.rwmtx.RUnlock()

	// Don't break logging. Just blackhole everything and be done. Can't send
	// everything to defaults since stdout might be closed, or if it's not,
	// too much logging could cause everything to block, which just defeats
	// the purpose.
	l.Reconfigure(Config{
		Outputs: map[string]*OutputConfig{
			"blackhole": &OutputConfig{
				Which: "blackhole",
				Level: Fatal,
			},
		},
		Modules: map[string]*ModuleConfig{
			"": &ModuleConfig{
				Outputs: []string{"blackhole"},
				Level:   Fatal,
			},
		},
	})

	// So this is kinda sketchy: finalizers for each output need to be called so that we can be sure they're done. Finalizers are only called during GCs... Yeah, loop on GC :(
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(time.Millisecond * 50):
				runtime.GC()
			case <-done:
				return
			}
		}
	}()

	wg.Wait()
	close(done)
}

// Rotate causes all outputters to rotate their files, if they have any. When
// using an external log rotator (eg. logrotated), this is what you're looking
// for to use in postrotate.
func (l *Log) Rotate() error {
	l.rwmtx.RLock()
	defer l.rwmtx.RUnlock()

	es := cog.Errors{}

	for name, o := range l.outputs {
		es.Addf(o.Rotate(), `failed to Rotate output "%s"`, name)
	}

	return es.Error()
}

// Get a logger for the given module
func (l *Log) Get(name string) *Logger {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	pfx, name := modulePrefix(name)

	lg, ok := l.active[name]
	if !ok {
		lg = &logger{
			l:   l,
			pfx: pfx,
			key: name,
		}

		lg.updateModule(l.modules)

		l.active[name] = lg
	}

	return newLogger(lg)
}
