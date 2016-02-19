package statc

import (
	"fmt"

	"github.com/thatguystone/cog"
	"github.com/thatguystone/cog/cio/eio"
	"github.com/thatguystone/cog/clog"
)

type output struct {
	out eio.Producer
	fmt Formatter
	log *clog.Logger
}

func newOutput(
	cfg OutputConfig,
	log *clog.Logger,
	exit *cog.GExit) (o *output, err error) {

	o = &output{
		log: log.Get(fmt.Sprintf("output.%s+%s",
			EscapePath(cfg.Prod),
			EscapePath(cfg.Fmt))),
	}

	o.fmt, err = newFormatter(cfg.Fmt, cfg.FmtArgs)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create formatter %s: %v",
			cfg.Fmt, err)
	}

	o.out, err = eio.NewProducer(cfg.Prod, cfg.ProdArgs)

	if err == nil {
		exit.Add(1)
		go o.run(exit)
	}

	return
}

func (o *output) send(snap Snapshot) {
	b, err := o.fmt.FormatSnap(snap)
	if err != nil {
		o.logErr(fmt.Errorf("format error: %v", err))
		return
	}

	o.out.Produce(b)
}

func (o *output) logErr(err error) {
	o.log.Errorf("%v", err)
}

func (o *output) run(exit *cog.GExit) {
	defer exit.Done()

	log := func(err error) {
		o.logErr(fmt.Errorf("producer error: %v", err))
	}

	errs := o.out.Errs()
	for {
		select {
		case err := <-errs:
			if err != nil {
				log(err)
			}

		case <-exit.C:
			es := o.out.Close()
			if !es.Empty() {
				log(es.Error())
			}
			return
		}
	}
}
