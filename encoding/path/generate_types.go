package path

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

const genHead = commentHead + `
package %s

import "github.com/thatguystone/cog/encoding/path"

`

type genMarshal struct {
	io.Writer
	pkgPath string
}

type genUnmarshal struct {
	io.Writer
	pkgPath string
	vars    *bytes.Buffer
}

var (
	marshalerType   = reflect.TypeOf((*Marshaler)(nil)).Elem()
	unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
)

func GenerateFromTypes(out, pkg string, types ...interface{}) error {
	f, err := os.Create(out)

	if err == nil {
		defer f.Close()
		err = generateFromTypesInto(f, pkg, types...)
	}

	return err
}

func generateFromTypesInto(w io.Writer, pkg string, types ...interface{}) error {
	_, err := fmt.Fprintf(w, genHead, pkg)

	if err == nil {
		vars := &bytes.Buffer{}

		gm := genMarshal{Writer: w}
		for _, t := range types {
			gum := genUnmarshal{
				Writer: w,
				vars:   vars,
			}

			err = gm.gen(t)
			if err == nil {
				err = gum.gen(t)
			}

			if err != nil {
				break
			}
		}

		if err == nil && vars.Len() > 0 {
			_, err = fmt.Fprintf(w, "var(\n%s\n)", vars.Bytes())
		}
	}

	return err
}

func (g *genMarshal) gen(t interface{}) (err error) {
	rt, _ := followPtrs(reflect.TypeOf(t))

	if g.pkgPath == "" {
		g.pkgPath = rt.PkgPath()
	} else if rt.PkgPath() != g.pkgPath {
		err = fmt.Errorf("may not mix package paths: have %s and %s",
			g.pkgPath,
			rt.PkgPath())
	}

	if err == nil {
		_, err = fmt.Fprintf(g,
			"func (v %s) MarshalPath(e path.Encoder) path.Encoder {\n",
			rt.Name())
	}

	if err == nil {
		err = g.genNamed("v", rt)
	}

	if err == nil {
		_, err = fmt.Fprintf(g, "	return e\n}\n\n")
	}

	return
}

func (g *genMarshal) genNamed(name string, rt reflect.Type) (err error) {
	brt, _ := followPtrs(rt)

	kind := brt.Kind()
	switch kind {
	case reflect.Struct:
		err = g.writeStruct(name, rt)

	case reflect.Array:
		err = g.writeArray(name, rt)

	case reflect.Bool,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		err = g.writeSimple(name, rt)

	default:
		err = fmt.Errorf("unsupported marshal kind: %s", kind)
	}

	return
}

func (g *genMarshal) writeStruct(name string, rt reflect.Type) (err error) {
	rt, depth := followPtrs(rt)
	if depth > 0 {
		name = fmt.Sprintf("(%s%s)", strings.Repeat("*", depth), name)
	}

	n := rt.NumField()
	for i := 0; err == nil && i < n; i++ {
		f := rt.Field(i)
		ft := f.Type
		name := fmt.Sprintf("%s.%s", name, f.Name)

		if ft == staticType {
			path := f.Tag.Get("path")
			_, err = fmt.Fprintf(g,
				"	e.B = append(e.B, \"%s\"...)\n"+
					"	e = e.EmitSep()\n",
				path)
			continue
		}

		if f.PkgPath != "" { // unexported
			continue
		}

		if ft.Implements(marshalerType) {
			_, err = fmt.Fprintf(g, "	e = (%s).MarshalPath(e)\n", name)
			continue
		}

		bft, _ := followPtrs(ft)
		kind := bft.Kind()
		switch kind {
		case reflect.Struct:
			// If an anonymous struct, need to emit it
			if f.Type.Name() == "" {
				g.writeStruct(name, ft)
			} else {
				_, err = fmt.Fprintf(g, "	e = e.Marshal(%s)\n", name)
			}

		case reflect.Array:
			err = g.writeArray(name, ft)

		case reflect.Bool,
			reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128,
			reflect.String:
			err = g.writeSimple(name, ft)

		default:
			err = fmt.Errorf("unsupported marshal kind: %s", kind)
		}
	}

	return
}

func (g *genMarshal) writeArray(name string, rt reflect.Type) (err error) {
	et := rt.Elem()

	switch {
	case et.Kind() == reflect.Int8 || et.Kind() == reflect.Uint8:
		_, err = fmt.Fprintf(g,
			"	e.B = append(e.B, %s...)\n"+
				"	e = e.EmitSep()\n",
			name)
	default:
		n := rt.Len()

		_, err = fmt.Fprintf(g,
			"	for i := 0; i < %d; i++ {\n",
			n)

		if err == nil {
			err = g.genNamed(fmt.Sprintf("%s[i]", name), et)
		}

		if err == nil {
			_, err = fmt.Fprintf(g, "	}\n")
		}
	}

	return
}

func (g *genMarshal) writeSimple(name string, rt reflect.Type) (err error) {
	bft, depth := followPtrs(rt)

	_, err = fmt.Fprintf(g, "	e = e.Emit%s(%s%s)\n",
		strings.Title(bft.Kind().String()),
		strings.Repeat("*", depth),
		name)
	return
}

func (g *genUnmarshal) gen(t interface{}) (err error) {
	rt, _ := followPtrs(reflect.TypeOf(t))

	_, err = fmt.Fprintf(g,
		"func (v *%s) UnmarshalPath(d path.Decoder) path.Decoder {\n",
		rt.Name())

	if err == nil {
		err = g.genNamed("v", rt, true)
	}

	if err == nil {
		_, err = fmt.Fprintf(g, "	return d\n}\n\n")
	}

	return
}

func (g *genUnmarshal) genNamed(name string, rt reflect.Type, guard bool) (err error) {
	brt, _ := followPtrs(rt)

	kind := brt.Kind()
	switch kind {
	case reflect.Struct:
		err = g.writeStruct(name, rt, guard)

	case reflect.Array:
		err = g.writeArray(name, rt, guard)

	case reflect.Bool,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		err = g.writeSimple(name, rt, guard)

	default:
		err = fmt.Errorf("unsupported marshal kind: %s", kind)
	}

	return
}

func (g *genUnmarshal) writeStruct(name string, rt reflect.Type, guard bool) (err error) {
	rt, depth := followPtrs(rt)
	if depth > 0 {
		name = fmt.Sprintf("(%s%s)", strings.Repeat("*", depth), name)
	}

	n := rt.NumField()
	for i := 0; err == nil && i < n; i++ {
		f := rt.Field(i)
		ft := f.Type

		name := fmt.Sprintf("%s.%s", name, f.Name)

		if ft == staticType {
			tagVarName := fmt.Sprintf("cogTag%s%d", rt.Name(), i)
			err = g.guard(fmt.Sprintf("d = d.ExpectTagBytes(%s)", tagVarName), guard)

			fmt.Fprintf(g.vars, "	%s = []byte(\"%s\")\n",
				tagVarName,
				f.Tag.Get("path"))

			continue
		}

		if f.PkgPath != "" { // unexported
			continue
		}

		if ft.Implements(unmarshalerType) {
			err = g.guard(fmt.Sprintf("d = (%s).UnmarshalPath(d)", name), guard)
			continue
		}

		bft, _ := followPtrs(ft)

		kind := bft.Kind()
		switch kind {
		case reflect.Struct:
			// If an anonymous struct, need to emit it
			if f.Type.Name() == "" {
				g.writeStruct(name, ft, guard)
			} else {
				err = g.guard(fmt.Sprintf("d = d.Unmarshal(&%s)", name), guard)
			}

		case reflect.Array:
			err = g.writeArray(name, ft, guard)

		case reflect.Bool,
			reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128,
			reflect.String:
			err = g.writeSimple(name, rt, guard)

		default:
			err = fmt.Errorf("unsupported marshal kind: %s", kind)
		}

		guard = true
	}

	return
}

func (g *genUnmarshal) writeArray(name string, rt reflect.Type, guard bool) (err error) {
	et := rt.Elem()

	switch {
	case et.Kind() == reflect.Int8 || et.Kind() == reflect.Uint8:
		err = g.guard(fmt.Sprintf("d = d.ExpectByteArray(%s[:])", name), true)

	default:
		n := rt.Len()

		_, err = fmt.Fprintf(g,
			"	for i := 0; d.Err == nil && i < %d; i++ {\n",
			n)

		if err == nil {
			err = g.genNamed(fmt.Sprintf("%s[i]", name), et, false)
		}

		if err == nil {
			_, err = fmt.Fprintf(g, "	}\n")
		}
	}

	return
}

func (g *genUnmarshal) writeSimple(name string, rt reflect.Type, guard bool) error {
	bft, depth := followPtrs(rt)

	_, err := fmt.Fprintf(g, "	e = e.Emit%s(&(%s%s))\n",
		strings.Title(bft.Kind().String()),
		strings.Repeat("*", depth),
		name)
	return err
}

func (g *genUnmarshal) guard(stmt string, guard bool) (err error) {
	stmt = strings.Join(strings.Split(stmt, "\n"), "\n\t\t")

	if guard {
		_, err = fmt.Fprintf(g,
			"	if d.Err == nil {\n"+
				"		%s\n"+
				"	}\n",
			stmt)
	} else {
		_, err = fmt.Fprintf(g, "		%s\n", stmt)
	}

	return
}

func followPtrs(rt reflect.Type) (reflect.Type, int) {
	depth := 0

	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		depth++
	}

	return rt, depth
}
