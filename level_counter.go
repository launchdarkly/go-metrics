package metrics

import "sync/atomic"

// Counters hold an int64 value that can be incremented and decremented.
type LevelCounter interface {
  Count() int64
  Dec(int64)
  Inc(int64)
  Snapshot() LevelCounter
}

// GetOrRegisterCounter returns an existing Counter or constructs and registers
// a new StandardCounter.
func GetOrRegisterLevelCounter(name string, r Registry) Counter {
  if nil == r {
    r = DefaultRegistry
  }
  return r.GetOrRegister(name, NewCounter).(Counter)
}

// NewCounter constructs a new StandardCounter.
func NewLevelCounter() LevelCounter {
  if UseNilMetrics {
    return NilLevelCounter{}
  }
  return &StandardLevelCounter{StandardCounter{0}}
}

// NewRegisteredCounter constructs and registers a new StandardLevelCounter.
func NewRegisteredLevelCounter(name string, r Registry) LevelCounter {
  c := NewLevelCounter()
  if nil == r {
    r = DefaultRegistry
  }
  r.Register(name, c)
  return c
}

// CounterSnapshot is a read-only copy of another Counter.
type LevelCounterSnapshot int64

// Count returns the count at the time the snapshot was taken.
func (c LevelCounterSnapshot) Count() int64 { return int64(c) }

// Dec panics.
func (LevelCounterSnapshot) Dec(int64) {
  panic("Dec called on a LevelCounterSnapshot")
}

// Inc panics.
func (LevelCounterSnapshot) Inc(int64) {
  panic("Inc called on a LevelCounterSnapshot")
}

// Snapshot returns the snapshot.
func (c LevelCounterSnapshot) Snapshot() LevelCounter { return c }

// NilCounter is a no-op Counter.
type NilLevelCounter struct {
  NilCounter
}

// Dec is a no-op.
func (NilLevelCounter) Dec(i int64) {}

// Snapshot is a no-op.
func (NilLevelCounter) Snapshot() LevelCounter { return NilLevelCounter{} }

// NilCounter is a no-op Counter.
type StandardLevelCounter struct {
  StandardCounter
}

// Dec decrements the counter by the given amount.
func (c *StandardLevelCounter) Dec(i int64) {
  atomic.AddInt64(&c.count, -i)
}

// Snapshot returns a read-only copy of the counter.
func (c *StandardLevelCounter) Snapshot() LevelCounter {
  return LevelCounterSnapshot(c.Count())
}

