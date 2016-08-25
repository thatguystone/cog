package statc

import (
	"sort"

	"github.com/iheartradio/cog"
)

// A Snapshot is a slice of stats sorted by name.
type Snapshot []Stat

// A Stat is a single stat value
type Stat struct {
	// String here, not Name: once in a snapshot, name should already have
	// been checked
	Name string
	Val  interface{}
}

// A Snapshotter takes a snapshot of itself, resets as necessary, and pushes
// the snapshotted value(s) to the given Adder.
type Snapshotter interface {
	Snapshot(a Adder)
}

type snapshotter struct {
	name    Name
	snapper Snapshotter
}

// A Name is your ticket to naming a stat. Names are, by default, prefixed
// with the *S's prefix from whence they came. They may also be appended to to
// create new Names.
//
// You should keep this around whenever you create a new Snapshotter since
// you'll need it to add the stat.
//
// This is really just a wrapper for a string, but when used right, it forces
// proper path caching and discourages generating garbage.
type Name struct {
	s  string
	ok bool
}

// An Adder is used to add stats to a snapshot. If no name is given, (that is
// Name{}), then the stat's name is used. If any other name is given, that
// will be used instead.
type Adder interface {
	AddBool(name Name, val bool)
	AddFloat(name Name, val float64)
	AddInt(name Name, val int64)
	AddString(name Name, val string)
}

type adder struct {
	s    *Snapshot
	name Name
}

// Get gets the Stat with the given name
func (s Snapshot) Get(name Name) Stat {
	for _, st := range s {
		if st.Name == name.Str() {
			return st
		}
	}

	return Stat{}
}

// Take takes a snapshot of a snapshotter
func (s *Snapshot) Take(name Name, sr Snapshotter) {
	a := adder{
		s:    s,
		name: name,
	}

	sr.Snapshot(a)
}

// Add a new value to this snapshot. Must be one of [bool, int64, float64,
// string].
func (s *Snapshot) Add(name Name, val interface{}) {
	s.add(name.Str(), val)
}

// Dup makes a copy of the snapshot
func (s Snapshot) Dup() (c Snapshot) {
	c = make(Snapshot, len(s))
	copy(c, s)
	return
}

func (s *Snapshot) add(name string, val interface{}) {
	l := len(*s)
	i := sort.Search(l, func(i int) bool {
		return (*s)[i].Name >= name
	})

	*s = append(*s, Stat{})

	copy((*s)[i+1:], (*s)[i:])
	(*s)[i] = Stat{
		Name: name,
		Val:  val,
	}
}

func newName(s string) Name {
	return Name{
		s:  s,
		ok: true,
	}
}

// Join combines this name with the given parts, escapes each part, and
// returns the new Name
func (n Name) Join(parts ...string) Name {
	return n.Append(JoinPath("", parts...))
}

// Append combines this name with the given name, without escaping anything
func (n Name) Append(name string) Name {
	return Name{
		s:  JoinNoEscape(n.s, name),
		ok: n.ok,
	}
}

// Str returns the name as a string
func (n Name) Str() string {
	cog.Assert(n.ok, "this name wasn't created from an *S. bad bad.")
	return n.s
}

func (a adder) add(name Name, val interface{}) {
	if name.s == "" {
		name = a.name
	}

	a.s.Add(name, val)
}

func (a adder) AddBool(name Name, val bool) {
	a.add(name, val)
}

func (a adder) AddFloat(name Name, val float64) {
	a.add(name, val)
}

func (a adder) AddInt(name Name, val int64) {
	a.add(name, val)
}

func (a adder) AddString(name Name, val string) {
	a.add(name, val)
}
