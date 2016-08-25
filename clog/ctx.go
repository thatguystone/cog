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
	"github.com/iheartradio/cog"
	"github.com/iheartradio/cog/cio/eio"
)

// Ctx provides access to the logging facilities
type Ctx struct {
	rwmtx sync.RWMutex
	logState

	mtx    sync.Mutex
	active map[string]*dLog
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
func New(cfg Config) (ctx *Ctx, err error) {
	ctx = &Ctx{
		active: map[string]*dLog{},
	}

	err = ctx.Reconfigure(cfg)
	if err != nil {
		ctx = nil
	}

	return
}

// NewFromFile creates a new Log, configured from the given file. The file type
// is determined by the extension.
func NewFromFile(path string) (ctx *Ctx, err error) {
	f, err := os.Open(path)

	if err == nil {
		defer f.Close()

		ext := filepath.Ext(strings.ToLower(path))
		switch ext {
		case ".json":
			ctx, err = NewFromJSONReader(f)

		default:
			err = fmt.Errorf("unsupported config file type: %s", ext)
		}
	}

	return
}

// NewFromJSONReader creates a new Log configured from JSON in the given
// reader.
func NewFromJSONReader(r io.Reader) (ctx *Ctx, err error) {
	d := json.NewDecoder(r)

	cfg := Config{}
	err = d.Decode(&cfg)

	if err == nil {
		ctx, err = New(cfg)
	}

	return
}

// Reconfigure reconfigures the entire logging system from the ground up. All
// active loggers are affected immediately, and all changes are applied
// atomically. If reconfiguration fails, the previous configuration remains.
func (ctx *Ctx) Reconfigure(cfg Config) error {
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
		out, err := newOutput(ocfg, ctx.Get("clog"), wg)
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
		ctx.rwmtx.Lock()
		defer ctx.rwmtx.Unlock()

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

		for _, lg := range ctx.active {
			lg.updateModule(tmp.modules)
		}

		ctx.logState = tmp
	}

	return err
}

// ReconfigureFromJSONReader reconfigures the Log from the JSON in the given
// reader.
func (ctx *Ctx) ReconfigureFromJSONReader(r io.Reader) error {
	d := json.NewDecoder(r)

	cfg := Config{}
	err := d.Decode(&cfg)

	if err == nil {
		err = ctx.Reconfigure(cfg)
	}

	return err
}

// ReconfigureFromFile reconfigures the Log from the given file. The file type
// is determined by the extension.
func (ctx *Ctx) ReconfigureFromFile(path string) error {
	f, err := os.Open(path)

	if err == nil {
		defer f.Close()

		ext := filepath.Ext(strings.ToLower(path))
		switch ext {
		case ".json":
			err = ctx.ReconfigureFromJSONReader(f)

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
func (ctx *Ctx) Flush() error {
	ctx.rwmtx.RLock()
	ls := ctx.logState
	ctx.rwmtx.RUnlock()

	// A reconfigure with the current options should work just fine. Make all
	// old outputs exit, which causes a flush.
	err := ctx.Reconfigure(ls.cfg)
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
func (ctx *Ctx) Rotate() error {
	ctx.rwmtx.RLock()
	defer ctx.rwmtx.RUnlock()

	es := cog.Errors{}

	for name, o := range ctx.outputs {
		es.Addf(o.Rotate(), `failed to Rotate output "%s"`, name)
	}

	return es.Error()
}

// Get a Log for the given module
func (ctx *Ctx) Get(name string) *Log {
	ctx.mtx.Lock()
	defer ctx.mtx.Unlock()

	pfx, name := modulePrefix(name)

	l, ok := ctx.active[name]
	if !ok {
		l = &dLog{
			mLog: &mLog{
				ctx: ctx,
				pfx: pfx,
				key: name,
				stats: Stats{
					Module: name,
				},
			},
		}

		if ctx.modules != nil {
			l.updateModule(ctx.modules)
		}

		ctx.active[name] = l
	}

	return newLogger(l)
}

// Stats gets the log stats since the last time this was called.
//
// Don't call this directly; it's meant for use with statc.
func (ctx *Ctx) Stats() (ss []Stats) {
	ctx.mtx.Lock()
	defer ctx.mtx.Unlock()

	for _, l := range ctx.active {
		ss = append(ss, l.stats.flush())
	}

	return
}
