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
	"github.com/thatguystone/cog/cio/eio"
)

// Log provides access to the logging facilities
type Log struct {
	rwmtx sync.RWMutex
	logState

	mtx    sync.Mutex
	active map[string]*logger
}

type logState struct {
	cfg     Config
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
		cfg:     cfg,
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
			Prod: "file",
			ProdArgs: eio.Args{
				"path": cfg.File,
			},
			Fmt:   "json",
			Level: Info,
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
			Prod:  "stdout",
			Fmt:   "human",
			Level: Info,
		}

		cfg.Modules[""] = &ModuleConfig{
			Outputs: []string{defaultTermOutputName},
			Level:   Info,
		}
	}

	es := cog.Errors{}
	for name, ocfg := range cfg.Outputs {
		out, err := newOutput(ocfg, l.Get("clog"), wg)
		if err != nil {
			es.Addf(err, `while creating output "%s"`, name)
			continue
		}

		tmp.outputs[name] = out
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

// Flush can be used by servers that exit gracefully. It causes all Outputs to
// write anything pending. It blocks until done.
//
// If it returns an error, nothing was flushed.
func (l *Log) Flush() error {
	l.rwmtx.RLock()
	ls := l.logState
	l.rwmtx.RUnlock()

	// A reconfigure with the current options should work just fine. Make all
	// old outputs exit, which causes a flush.
	err := l.Reconfigure(ls.cfg)
	if err != nil {
		return err
	}

	// So this is kinda sketchy: finalizers for each output need to be
	// called so that we can be sure they're done. Finalizers are only
	// called during GCs... Yeah, loop on GC :(
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

	ls.wg.Wait()
	close(done)

	return nil
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
			stats: Stats{
				Module: name,
			},
		}

		if l.modules != nil {
			lg.updateModule(l.modules)
		}

		l.active[name] = lg
	}

	return newLogger(lg)
}

// Stats gets the log stats since the last time this was called.
//
// Don't call this directly; it's meant for use with statc.
func (l *Log) Stats() (ss []Stats) {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	for _, lg := range l.active {
		ss = append(ss, lg.stats.flush())
	}

	return
}
