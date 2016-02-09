package stats

// A Snapshot is a slice of stats sorted by name.
type Snapshot []Stat

// A Stat is a single stat value
type Stat struct {
	Name string
	Val  interface{}
}

// A Snapshotter takes a snapshot of itself, resets as necessary, and pushes
// the snapshotted value(s) to the given Added.
type Snapshotter interface {
	Snapshot(a Adder)
}

type snapshotter struct {
	name    string
	snapper Snapshotter
}

// An Adder is used to add stats to a snapshot. If no name is given, (that is,
// name=""), then the stat's name is used.
type Adder interface {
	AddBool(name string, val bool)
	AddFloat(name string, val float64)
	AddInt(name string, val int64)
	AddString(name, val string)
}

type adder struct {
	s    *Snapshot
	name string
}

// Get gets the Stat with the given name
func (s Snapshot) Get(name string) Stat {
	for _, st := range s {
		if st.Name == name {
			return st
		}
	}

	return Stat{}
}

// Take takes a snapshot of a snapshotter
func (s *Snapshot) Take(name string, sr Snapshotter) {
	a := adder{
		s:    s,
		name: name,
	}

	sr.Snapshot(a)
}

// Add a new value to this snapshot. Must be one of [bool, int64, float64,
// string].
func (s *Snapshot) Add(name string, val interface{}) {
	*s = append(*s, Stat{
		Name: name,
		Val:  val,
	})
}

func (a adder) add(name string, val interface{}) {
	if name == "" {
		name = a.name
	}

	a.s.Add(name, val)
}

func (a adder) AddBool(name string, val bool) {
	a.add(name, val)
}

func (a adder) AddFloat(name string, val float64) {
	a.add(name, val)
}

func (a adder) AddInt(name string, val int64) {
	a.add(name, val)
}

func (a adder) AddString(name, val string) {
	a.add(name, val)
}
