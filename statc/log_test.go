package statc

import "testing"

func TestLogStats(t *testing.T) {
	c, st := newTest(t, nil)
	defer st.exit.Exit()

	st.AddLog("log", st.log)

	lg := st.log.Get("fun")
	lg.Debug("123")
	for i := 0; i < 5; i++ {
		lg.Info("123")
	}

	lg = st.log.Get("mean")
	lg.Warn("123")
	for i := 0; i < 5; i++ {
		lg.Error("123")
	}

	snap := st.snapshot()
	for _, st := range snap {
		c.Logf("%s = %v", st.Name.Str(), st.Val)
	}

	c.Equal(snap.Get(st.Names("log", "fun", "debug")).Val.(int64), 1)
	c.Equal(snap.Get(st.Names("log", "fun", "info")).Val.(int64), 5)
	c.Equal(snap.Get(st.Names("log", "fun", "warn")).Val.(int64), 0)
	c.Equal(snap.Get(st.Names("log", "fun", "error")).Val.(int64), 0)

	c.Equal(snap.Get(st.Names("log", "mean", "debug")).Val.(int64), 0)
	c.Equal(snap.Get(st.Names("log", "mean", "info")).Val.(int64), 0)
	c.Equal(snap.Get(st.Names("log", "mean", "warn")).Val.(int64), 1)
	c.Equal(snap.Get(st.Names("log", "mean", "error")).Val.(int64), 5)
}
