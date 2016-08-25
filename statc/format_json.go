package statc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pquerna/ffjson/fflib/v1"
	"github.com/iheartradio/cog/cio/eio"
	"github.com/iheartradio/cog/unsafec"
)

// JSONFormat formats snapshots as JSON
type JSONFormat struct {
	Args struct {
		// If the JSON should be pretty-printed
		Pretty bool
	}
}

type jsonState struct {
	bytes.Buffer
	err       error
	snap      Snapshot
	needComma bool
}

var (
	bTrue  = []byte("true")
	bFalse = []byte("false")
)

func init() {
	RegisterFormatter("json",
		func(args eio.Args) (Formatter, error) {
			jf := JSONFormat{}
			err := args.ApplyTo(&jf.Args)
			return jf, err
		})
}

// FormatSnap implements Formatter.FormatSnap
func (j JSONFormat) FormatSnap(snap Snapshot) ([]byte, error) {
	js := jsonState{
		snap: snap,
	}

	js.dump("")

	err := js.err
	if err == nil && j.Args.Pretty {
		b2 := bytes.Buffer{}

		err = json.Indent(&b2, js.Bytes(), "", "\t")
		if err == nil {
			js.Buffer = b2
		}
	}

	return js.Bytes(), err
}

// MimeType implements Formatter.MimeType
func (JSONFormat) MimeType() string {
	return "application/json"
}

func (js *jsonState) dump(prefix string) {
	js.open()

	for len(js.snap) > 0 && js.err == nil {
		stat := js.snap[0]

		if !strings.HasPrefix(stat.Name, prefix) {
			break
		}

		name := stat.Name[len(prefix):]

		doti := strings.IndexByte(name, '.')
		if doti != -1 {
			js.key(name[:doti])
			js.dump(stat.Name[:len(prefix)+doti+1])
			continue
		}

		js.snap = js.snap[1:]

		var vb []byte

		switch v := stat.Val.(type) {
		case string:
			vb, js.err = json.Marshal(stat.Val)

		case int64:
			vb = unsafec.Bytes(strconv.FormatInt(v, 10))

		case float64:
			vb = unsafec.Bytes(strconv.FormatFloat(v, 'f', -1, 64))

		case bool:
			if v {
				vb = bTrue
			} else {
				vb = bFalse
			}

		default:
			js.err = fmt.Errorf("unrecognized type: %T", v)
		}

		js.key(name)
		js.val(vb)
	}

	js.close()

	return
}

func (js *jsonState) open() {
	js.comma()
	if js.err == nil {
		js.WriteByte('{')
	}
}

func (js *jsonState) close() {
	if js.err == nil {
		js.WriteByte('}')
	}
	js.needComma = true
}

func (js *jsonState) key(k string) {
	js.comma()
	if js.err == nil {
		v1.WriteJson(js, unsafec.Bytes(k))
		js.WriteByte(':')
	}
}

func (js *jsonState) val(b []byte) {
	js.write(b)
	js.needComma = true
}

func (js *jsonState) comma() {
	if js.err == nil && js.needComma {
		js.WriteByte(',')
		js.needComma = false
	}
}

func (js *jsonState) write(b []byte) {
	if js.err == nil {
		_, js.err = js.Write(b)
	}
}
