package clog

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/iheartradio/cog/cio/eio"
)

type privOutput struct {
	eio.Producer
	Formatter
	data  Data
	l     *Log
	wg    *sync.WaitGroup
	filts filterSlice
}

// outputs are tricky business: they can be used without a lock in many
// goroutines, so they're not safe to close until no one is using them: a
// finalizer is required.
type output struct {
	*privOutput
}

func newOutput(oc *OutputConfig, l *Log, wg *sync.WaitGroup) (o *output, err error) {
	po := &privOutput{
		data: Data{
			"prod":     oc.Prod,
			"prodArgs": fmt.Sprintf("%+v", oc.ProdArgs),
			"fmt":      oc.Fmt,
			"fmtArgs":  fmt.Sprintf("%+v", oc.FmtArgs),
		},
		l:  l,
		wg: wg,
	}

	wg.Add(1)
	defer func() {
		if err != nil {
			po.exit()
			o = nil
		}
	}()

	po.Formatter, err = newFormatter(oc.Fmt, oc.FmtArgs)
	if err != nil {
		return
	}

	if oc.ProdArgs == nil {
		oc.ProdArgs = eio.Args{}
	}

	oc.ProdArgs["MimeType"] = po.Formatter.MimeType()

	po.Producer, err = eio.NewProducer(oc.Prod, oc.ProdArgs)
	if err != nil {
		return
	}

	po.filts, err = newFilters(oc.Level, oc.Filters)
	if err != nil {
		return
	}

	wg.Add(1)
	go po.monitor()

	o = &output{po}
	runtime.SetFinalizer(o, finalizeOutput)

	return
}

func finalizeOutput(o *output) {
	go o.exit()
}

// May only be used by *privOutput itself.
func (po *privOutput) exit() {
	if po.Producer != nil {
		errs := po.Producer.Close()
		if !errs.Empty() {
			po.logErr(errs.Error())
		}
	}

	for _, f := range po.filts {
		f.Exit()
	}

	po.wg.Done()
}

func (po *privOutput) logErr(err error) {
	if err != nil {
		po.l.LogEntry(Entry{
			Level:        Error,
			Depth:        1,
			Msg:          fmt.Sprintf("failed to write log entry: %v", err),
			Data:         po.data,
			ignoreErrors: true,
		})
	}
}

func (po *privOutput) monitor() {
	defer po.wg.Done()
	for err := range po.Errs() {
		po.logErr(err)
	}
}
