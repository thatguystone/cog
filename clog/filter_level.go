package clog

import "github.com/iheartradio/cog/cio/eio"

// LevelFilter is the filter used by the required "Level" argument for both
// Modules and Outputs and is typically not used directly.
//
// Example:
//
//    Filters: []FilterConfig{
//        FilterConfig{
//            Which: "Level",
//            Args: eio.Args{
//                "level": clog.Info,
//            },
//        },
//    }
type LevelFilter struct {
	Args struct {
		// Don't log anything below this level
		Level Level
	}
}

const lvlFilterName = "Level"

func init() {
	RegisterFilter(lvlFilterName,
		func(a eio.Args) (Filter, error) {
			f := LevelFilter{}

			err := a.ApplyTo(&f.Args)
			if err != nil {
				return nil, err
			}

			return f, nil
		})
}

// Accept implements Filter.Accept()
func (lf LevelFilter) Accept(e Entry) bool {
	return e.Level >= lf.Args.Level
}

// Exit implements Filter.Exit()
func (lf LevelFilter) Exit() {}
