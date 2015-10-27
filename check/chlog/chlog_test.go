package chlog

import "testing"

func TestChlog(t *testing.T) {
	_, log := New(t)
	log.Get("test").Info("here")
}
