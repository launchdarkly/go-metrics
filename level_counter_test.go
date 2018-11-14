package metrics

import "testing"

func BenchmarkLevelCounter(b *testing.B) {
  c := NewLevelCounter()
  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    c.Inc(1)
  }
}

func TestLevelCounterDec1(t *testing.T) {
  c := NewLevelCounter()
  c.Dec(1)
  if count := c.Count(); -1 != count {
    t.Errorf("c.Count(): -1 != %v\n", count)
  }
}

func TestLevelCounterDec2(t *testing.T) {
  c := NewLevelCounter()
  c.Dec(2)
  if count := c.Count(); -2 != count {
    t.Errorf("c.Count(): -2 != %v\n", count)
  }
}

func TestLevelCounterInc1(t *testing.T) {
  c := NewLevelCounter()
  c.Inc(1)
  if count := c.Count(); 1 != count {
    t.Errorf("c.Count(): 1 != %v\n", count)
  }
}

func TestLevelCounterInc2(t *testing.T) {
  c := NewLevelCounter()
  c.Inc(2)
  if count := c.Count(); 2 != count {
    t.Errorf("c.Count(): 2 != %v\n", count)
  }
}

func TestLevelCounterSnapshot(t *testing.T) {
  c := NewLevelCounter()
  c.Inc(1)
  snapshot := c.Snapshot()
  c.Inc(1)
  if count := snapshot.Count(); 1 != count {
    t.Errorf("c.Count(): 1 != %v\n", count)
  }
}

func TestLevelCounterZero(t *testing.T) {
  c := NewCounter()
  if count := c.Count(); 0 != count {
    t.Errorf("c.Level(): 0 != %v\n", count)
  }
}

func TestGetOrRegisterLevelCounter(t *testing.T) {
  r := NewRegistry()
  NewRegisteredCounter("foo", r).Inc(47)
  if c := GetOrRegisterLevelCounter("foo", r); 47 != c.Count() {
    t.Fatal(c)
  }
}
