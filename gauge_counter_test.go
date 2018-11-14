package metrics

import "testing"

func BenchmarkGaugeCounter(b *testing.B) {
  c := NewGaugeCounter()
  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    c.Inc(1)
  }
}

func TestGaugeCounterDec1(t *testing.T) {
  c := NewGaugeCounter()
  c.Dec(1)
  if count := c.Count(); -1 != count {
    t.Errorf("c.Count(): -1 != %v\n", count)
  }
}

func TestGaugeCounterDec2(t *testing.T) {
  c := NewGaugeCounter()
  c.Dec(2)
  if count := c.Count(); -2 != count {
    t.Errorf("c.Count(): -2 != %v\n", count)
  }
}

func TestGaugeCounterInc1(t *testing.T) {
  c := NewGaugeCounter()
  c.Inc(1)
  if count := c.Count(); 1 != count {
    t.Errorf("c.Count(): 1 != %v\n", count)
  }
}

func TestGaugeCounterInc2(t *testing.T) {
  c := NewGaugeCounter()
  c.Inc(2)
  if count := c.Count(); 2 != count {
    t.Errorf("c.Count(): 2 != %v\n", count)
  }
}

func TestGaugeCounterSnapshot(t *testing.T) {
  c := NewGaugeCounter()
  c.Inc(1)
  snapshot := c.Snapshot()
  c.Inc(1)
  if count := snapshot.Count(); 1 != count {
    t.Errorf("c.Count(): 1 != %v\n", count)
  }
}

func TestGaugeCounterZero(t *testing.T) {
  c := NewCounter()
  if count := c.Count(); 0 != count {
    t.Errorf("c.Count(): 0 != %v\n", count)
  }
}

func TestGetOrRegisterGaugeCounter(t *testing.T) {
  r := NewRegistry()
  NewRegisteredCounter("foo", r).Inc(47)
  if c := GetOrRegisterGaugeCounter("foo", r); 47 != c.Count() {
    t.Fatal(c)
  }
}
