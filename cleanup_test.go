// +build go1.14

package quicktest_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

// This file defines tests that are only valid since the Cleanup
// method was added in Go 1.14.

func TestCCleanup(t *testing.T) {
	c := qt.New(t)
	cleanups := 0
	c.Run("defer", func(c *qt.C) {
		c.Cleanup(func() {
			cleanups++
		})
	})
	c.Assert(cleanups, qt.Equals, 1)
}

func TestCDeferWithoutDone(t *testing.T) {
	c := qt.New(t)
	tc := &testingTWithCleanup{
		TB:      t,
		cleanup: func() {},
	}
	c1 := qt.New(tc)
	c1.Defer(func() {})
	c1.Defer(func() {})
	c.Assert(tc.cleanup, qt.PanicMatches, `Done not called after Defer`)
}

func TestCDeferVsCleanupOrder(t *testing.T) {
	c := qt.New(t)
	var defers []string
	testDefer(c, func(c *qt.C) {
		c.Defer(func() {
			defers = append(defers, "defer-0")
		})
		c.Cleanup(func() {
			defers = append(defers, "cleanup-0")
		})
		c.Defer(func() {
			defers = append(defers, "defer-1")
		})
		c.Cleanup(func() {
			defers = append(defers, "cleanup-1")
		})
	})
	c.Assert(defers, qt.DeepEquals, []string{"defer-1", "defer-0", "cleanup-1", "cleanup-0"})
}

func TestCDeferInSubC(t *testing.T) {
	c := qt.New(t)
	var defers []int
	testDefer(c, func(c *qt.C) {
		c.Defer(func() {
			defers = append(defers, 0)
		})
		c2 := qt.New(c)
		c2.Defer(func() {
			defers = append(defers, 1)
		})
		c2.Done()
		c.Check(defers, qt.DeepEquals, []int{1})
		c.Defer(func() {
			defers = append(defers, 2)
		})
	})
	c.Assert(defers, qt.DeepEquals, []int{1, 2, 0})
}

type testingTWithCleanup struct {
	testing.TB
	cleanup func()
}

func (t *testingTWithCleanup) Cleanup(f func()) {
	oldCleanup := t.cleanup
	t.cleanup = func() {
		defer oldCleanup()
		f()
	}
}
