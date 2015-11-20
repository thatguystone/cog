package clog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tchap/go-patricia/patricia"
	"github.com/thatguystone/cog"
)

// Log provides access to the logging facilities
type Log struct {
	rwmtx   sync.RWMutex
	outputs map[string]*output
	modules *patricia.Trie // Items of type *module

	mtx    sync.Mutex
	active map[string]*logger
}

type module struct {
	pfx           patricia.Prefix
	parent        *module
	outs          []*output
	filts         *filters
	dontPropagate bool
}

type output struct {
	Outputter
	filts *filters
}

// New creates a new logger
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
	tmp := Log{
		outputs: map[string]*output{},
		modules: patricia.NewTrie(),
	}

	if cfg.Outputs == nil {
		cfg.Outputs = map[string]*ConfigOutput{}
	}

	if cfg.Modules == nil {
		cfg.Modules = map[string]*ConfigModule{}
	}

	if cfg.File != "" {
		cfg.Outputs[defaultConfigFileOutputName] = &ConfigOutput{
			Which: "jsonfile",
			Level: Info,
			Args: ConfigOutputArgs{
				"path": cfg.File,
			},
		}

		m, ok := cfg.Modules[""]
		if !ok {
			m = &ConfigModule{
				Level: Info,
			}
			cfg.Modules[""] = m
		}

		m.Outputs = append(m.Outputs, defaultConfigFileOutputName)
	}

	if len(cfg.Modules) == 0 {
		cfg.Outputs[defaultTermOutputName] = &ConfigOutput{
			Which: "term",
			Level: Info,
		}

		cfg.Modules[""] = &ConfigModule{
			Outputs: []string{defaultTermOutputName},
			Level:   Info,
		}
	}

	es := cog.Errors{}
	for name, ocfg := range cfg.Outputs {
		newOut, ok := regdOutputs[strings.ToLower(ocfg.Which)]
		if !ok {
			es.Add(fmt.Errorf(`in output "%s": output which="%s" does not exist`,
				name,
				ocfg.Which))
			continue
		}

		o, err := newOut(ocfg.Args)

		if err != nil {
			es.Addf(err, `while creating output "%s"`, name)
			continue
		}

		filts, err := newFilters(ocfg.Level, ocfg.Filters)
		if err != nil {
			es.Addf(err, `for output "%s", invalid filter config`, name)
			continue
		}

		tmp.outputs[name] = &output{
			Outputter: o,
			filts:     filts,
		}
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

		l.outputs = tmp.outputs
		l.modules = tmp.modules
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

// Reopen causes all outputters to reopen their files, if they have any. When
// using an external log rotator (eg. logrotated), this is what you're looking
// for to use in postrotate.
func (l *Log) Reopen() error {
	l.rwmtx.RLock()
	defer l.rwmtx.RUnlock()

	es := cog.Errors{}

	for name, o := range l.outputs {
		es.Addf(o.Reopen(), `failed to reopen output "%s"`, name)
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
