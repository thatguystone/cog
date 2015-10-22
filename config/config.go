// Package config implements pluggable JSON configuration for multiple modules.
package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/DisposaBoy/JsonConfigReader"
	"github.com/thatguystone/cog"
)

// Cfg handles getting config setup for everyone
type Cfg struct {
	// All loaded modules
	Modules Modules

	// For testing: change where to call for EC2 metadata
	ec2MetadataBase string
	ec2PrivIPv4     string
}

// Modules provides a shorter way of writing out the map necessary for listing
// all modules
type Modules map[string]Configer

// Configer provides a way for modules to get configured in a group with all
// other modules.
type Configer interface {
	// Validate everything
	Validate(*Cfg, *cog.Errors)
}

// NewConfiger creates an instance of a Configer
type NewConfiger func() Configer

var (
	mtx     sync.Mutex
	modules = map[string]NewConfiger{}
)

// Register adds a new function to the list of module configs. This function is
// called to create a new Configer with all default options set.
func Register(name string, ncfgr NewConfiger) {
	mtx.Lock()
	defer mtx.Unlock()

	if _, ok := modules[name]; ok {
		panic(fmt.Errorf("config module `%s` already exists", name))
	}

	modules[name] = ncfgr
}

// New creates a new instance of *Cfg with all registered modules
func New() *Cfg {
	cfg := &Cfg{
		Modules: Modules{},
	}

	mtx.Lock()
	defer mtx.Unlock()

	for name, cfgr := range modules {
		cfg.Modules[name] = cfgr()
	}

	return cfg
}

// LoadAndValidate loads the configuration file then validates it all
func (cfg *Cfg) LoadAndValidate(path string) error {
	f, err := os.Open(path)

	m := map[string]interface{}{}
	if err == nil {
		defer f.Close()

		// Loading directly into the Modules map doesn't work, so just make this
		// nice and generic: it loads from the file, the it Marshals and re-
		// Unmarshals into each individual config
		dec := json.NewDecoder(JsonConfigReader.New(f))
		err = dec.Decode(&m)
	}

	if err == nil {
		fillIn := func(p interface{}, i interface{}) error {
			out, err := json.Marshal(p)
			if err == nil {
				err = json.Unmarshal(out, i)
			}

			return err
		}

		err = fillIn(m, cfg)
		for name, cfger := range cfg.Modules {
			if err == nil {
				err = fillIn(m[name], cfger)
			}
		}
	}

	if err == nil {
		err = cfg.Validate()
	}

	return err
}

// Validate validates the current state of the config
func (cfg *Cfg) Validate() (err error) {
	es := cog.Errors{}
	defer func() {
		err = es.Error()
	}()

	for name, cfger := range cfg.Modules {
		cfger.Validate(cfg, es.Prefix("in "+name))
	}

	return
}

// ResolveListen validates the given address and resolves any special hostnames
// to actual addresses.
func (cfg *Cfg) ResolveListen(addr *string, es *cog.Errors) {
	// Just let it fail on listen
	if len(*addr) == 0 {
		es.Add(fmt.Errorf("no listen address given"))
		return
	}

	host, port, err := net.SplitHostPort(*addr)
	if err != nil {
		es.Add(fmt.Errorf("invalid listen address: %s: %v", *addr, err))
	} else if host == "ec2" {
		ip, err := cfg.GetEC2PrivIP()
		if err != nil {
			es.Add(fmt.Errorf("while getting ec2 metadata for %s: %v", *addr, err))
		} else {
			*addr = fmt.Sprintf("%s:%s", ip, port)
		}
	}
}
