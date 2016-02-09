package clog

// LevelFilter is the filter used by the required "Level" argument for both
// Modules and Outputs.
type LevelFilter struct {
	args struct {
		Level Level
	}
}

const lvlFilterName = "Level"

func init() {
	RegisterFilter(lvlFilterName,
		func(a ConfigArgs) (Filter, error) {
			f := LevelFilter{}

			err := a.ApplyTo(&f.args)
			if err != nil {
				return nil, err
			}

			return f, nil
		})
}

// Accept implements Filter.Accept()
func (lf LevelFilter) Accept(e Entry) bool {
	return e.Level >= lf.args.Level
}

// Exit implements Filter.Exit()
func (lf LevelFilter) Exit() {}
