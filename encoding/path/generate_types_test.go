package path

import (
	"bytes"
	"sync"
	"testing"

	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/cio"
	"github.com/thatguystone/cog/ctime"
)

type arrays struct {
	A [8]byte
	B [8]int8
	C [8]uint8
	D [8]uint32
	E [8][2]uint32
}

func TestGenerateFromTypes(t *testing.T) {
	c := check.New(t)

	b := bytes.Buffer{}
	err := generateFromTypesInto(&b, "test",
		Pie{},
		new(arrays),
		Everything{},
		new(EverythingPtr))
	c.MustNotError(err)

	s := b.String()
	c.Contains(s, "func (v Pie) MarshalPath(e path.Encoder) path.Encoder {")
	c.Contains(s, `e.B = append(e.B, "pie"...)`+"\n	e = e.EmitSep()")
	c.Contains(s, "func (v arrays) MarshalPath(e path.Encoder) path.Encoder {")

	c.Contains(s, "func (v *arrays) UnmarshalPath(d path.Decoder) path.Decoder {")

	// Exported fields shouldn't be around
	c.NotContains(s, "v.g")
}

func TestGenerateFromTypesErrors(t *testing.T) {
	c := check.New(t)

	b := bytes.Buffer{}
	err := generateFromTypesInto(&b, "test", []string{})
	c.Error(err)

	err = generateFromTypesInto(&b, "test",
		new(ctime.HumanDuration),
		Pie{})
	c.Error(err)
}

func TestGenerateFromTypesWriteErrors(t *testing.T) {
	c := check.New(t)

	buff := &bytes.Buffer{}
	err := generateFromTypesInto(buff, "test",
		Pie{},
		new(arrays),
		Everything{},
		new(EverythingPtr))
	c.MustNotError(err)

	wg := sync.WaitGroup{}
	wg.Add(buff.Len())
	for i := 0; i < buff.Len(); i++ {
		go func(max int) {
			defer wg.Done()

			out := &bytes.Buffer{}
			b := cio.LimitedWriter{
				W: out,
				N: int64(max),
			}

			err := generateFromTypesInto(&b, "test",
				Pie{},
				new(arrays),
				Everything{},
				new(EverythingPtr))
			c.Equal(max, out.Len())
			c.Error(err, "failed at max=%d", max, out.String())
		}(i)
	}

	wg.Wait()
}
