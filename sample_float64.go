package metrics

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// SampleFloat64s maintain a statistically-significant selection of values from
// a stream.
type SampleFloat64 interface {
	Clear()
	Count() int64
	Max() float64
	Mean() float64
	Min() float64
	Percentile(float64) float64
	Percentiles([]float64) []float64
	Size() int
	Snapshot() SampleFloat64
	StdDev() float64
	Sum() float64
	Update(float64)
	Values() []float64
	Variance() float64
}

// ExpDecaySampleFloat64 is an exponentially-decaying SampleFloat64 using a forward-decaying
// priority reservoir.  See Cormode et al's "Forward Decay: A Practical Time
// Decay Model for Streaming Systems".
//
// <http://dimacs.rutgers.edu/~graham/pubs/papers/fwddecay.pdf>
type ExpDecaySampleFloat64 struct {
	alpha         float64
	count         int64
	mutex         sync.Mutex
	reservoirSize int
	t0, t1        time.Time
	values        *expDecaySampleFloat64Heap
}

// NewExpDecaySampleFloat64 constructs a new exponentially-decaying SampleFloat64 with the
// given reservoir size and alpha.
func NewExpDecaySampleFloat64(reservoirSize int, alpha float64) SampleFloat64 {
	if UseNilMetrics {
		return NilSampleFloat64{}
	}
	s := &ExpDecaySampleFloat64{
		alpha:         alpha,
		reservoirSize: reservoirSize,
		t0:            time.Now(),
		values:        newExpDecaySampleFloat64Heap(reservoirSize),
	}
	s.t1 = s.t0.Add(rescaleThreshold)
	return s
}

// Clear clears all SampleFloat64s.
func (s *ExpDecaySampleFloat64) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.count = 0
	s.t0 = time.Now()
	s.t1 = s.t0.Add(rescaleThreshold)
	s.values.Clear()
}

// Count returns the number of SampleFloat64s recorded, which may exceed the
// reservoir size.
func (s *ExpDecaySampleFloat64) Count() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.count
}

// Max returns the maximum value in the SampleFloat64, which may not be the maximum
// value ever to be part of the SampleFloat64.
func (s *ExpDecaySampleFloat64) Max() float64 {
	return SampleFloat64Max(s.Values())
}

// Mean returns the mean of the values in the SampleFloat64.
func (s *ExpDecaySampleFloat64) Mean() float64 {
	return SampleFloat64Mean(s.Values())
}

// Min returns the minimum value in the SampleFloat64, which may not be the minimum
// value ever to be part of the SampleFloat64.
func (s *ExpDecaySampleFloat64) Min() float64 {
	return SampleFloat64Min(s.Values())
}

// Percentile returns an arbitrary percentile of values in the SampleFloat64.
func (s *ExpDecaySampleFloat64) Percentile(p float64) float64 {
	return SampleFloat64Percentile(s.Values(), p)
}

// Percentiles returns a slice of arbitrary percentiles of values in the
// SampleFloat64.
func (s *ExpDecaySampleFloat64) Percentiles(ps []float64) []float64 {
	return SampleFloat64Percentiles(s.Values(), ps)
}

// Size returns the size of the SampleFloat64, which is at most the reservoir size.
func (s *ExpDecaySampleFloat64) Size() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.values.Size()
}

// Snapshot returns a read-only copy of the SampleFloat64.
func (s *ExpDecaySampleFloat64) Snapshot() SampleFloat64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	vals := s.values.Values()
	values := make([]float64, len(vals))
	for i, v := range vals {
		values[i] = v.v
	}
	return &SampleFloat64Snapshot{
		count:  s.count,
		values: values,
	}
}

// StdDev returns the standard deviation of the values in the SampleFloat64.
func (s *ExpDecaySampleFloat64) StdDev() float64 {
	return SampleFloat64StdDev(s.Values())
}

// Sum returns the sum of the values in the SampleFloat64.
func (s *ExpDecaySampleFloat64) Sum() float64 {
	return SampleFloat64Sum(s.Values())
}

// Update SampleFloat64s a new value.
func (s *ExpDecaySampleFloat64) Update(v float64) {
	s.update(time.Now(), v)
}

// Values returns a copy of the values in the SampleFloat64.
func (s *ExpDecaySampleFloat64) Values() []float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	vals := s.values.Values()
	values := make([]float64, len(vals))
	for i, v := range vals {
		values[i] = v.v
	}
	return values
}

// Variance returns the variance of the values in the SampleFloat64.
func (s *ExpDecaySampleFloat64) Variance() float64 {
	return SampleFloat64Variance(s.Values())
}

// update SampleFloat64s a new value at a particular timestamp.  This is a method all
// its own to facilitate testing.
func (s *ExpDecaySampleFloat64) update(t time.Time, v float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.count++
	if s.values.Size() == s.reservoirSize {
		s.values.Pop()
	}
	s.values.Push(expDecaySampleFloat64{
		k: math.Exp(t.Sub(s.t0).Seconds()*s.alpha) / rand.Float64(),
		v: v,
	})
	if t.After(s.t1) {
		values := s.values.Values()
		t0 := s.t0
		s.values.Clear()
		s.t0 = t
		s.t1 = s.t0.Add(rescaleThreshold)
		for _, v := range values {
			v.k = v.k * math.Exp(-s.alpha*s.t0.Sub(t0).Seconds())
			s.values.Push(v)
		}
	}
}

// NilSampleFloat64 is a no-op SampleFloat64.
type NilSampleFloat64 struct{}

// Clear is a no-op.
func (NilSampleFloat64) Clear() {}

// Count is a no-op.
func (NilSampleFloat64) Count() int64 { return 0 }

// Max is a no-op.
func (NilSampleFloat64) Max() float64 { return 0 }

// Mean is a no-op.
func (NilSampleFloat64) Mean() float64 { return 0.0 }

// Min is a no-op.
func (NilSampleFloat64) Min() float64 { return 0 }

// Percentile is a no-op.
func (NilSampleFloat64) Percentile(p float64) float64 { return 0.0 }

// Percentiles is a no-op.
func (NilSampleFloat64) Percentiles(ps []float64) []float64 {
	return make([]float64, len(ps))
}

// Size is a no-op.
func (NilSampleFloat64) Size() int { return 0 }

// SampleFloat64 is a no-op.
func (NilSampleFloat64) Snapshot() SampleFloat64 { return NilSampleFloat64{} }

// StdDev is a no-op.
func (NilSampleFloat64) StdDev() float64 { return 0.0 }

// Sum is a no-op.
func (NilSampleFloat64) Sum() float64 { return 0 }

// Update is a no-op.
func (NilSampleFloat64) Update(v float64) {}

// Values is a no-op.
func (NilSampleFloat64) Values() []float64 { return []float64{} }

// Variance is a no-op.
func (NilSampleFloat64) Variance() float64 { return 0.0 }

// SampleFloat64Max returns the maximum value of the slice of float64.
func SampleFloat64Max(values []float64) float64 {
	if 0 == len(values) {
		return 0
	}
	var max float64 = math.MaxFloat64 * -1
	for _, v := range values {
		if max < v {
			max = v
		}
	}
	return max
}

// SampleFloat64Mean returns the mean value of the slice of float64.
func SampleFloat64Mean(values []float64) float64 {
	if 0 == len(values) {
		return 0.0
	}
	return float64(SampleFloat64Sum(values)) / float64(len(values))
}

// SampleFloat64Min returns the minimum value of the slice of int64.
func SampleFloat64Min(values []float64) float64 {
	if 0 == len(values) {
		return 0
	}
	var min float64 = math.MaxFloat64
	for _, v := range values {
		if min > v {
			min = v
		}
	}
	return min
}

// SampleFloat64Percentiles returns an arbitrary percentile of the slice of
// float64.
func SampleFloat64Percentile(values float64Slice, p float64) float64 {
	return SampleFloat64Percentiles(values, []float64{p})[0]
}

// SampleFloat64Percentiles returns a slice of arbitrary percentiles of the slice of
// float64.
func SampleFloat64Percentiles(values float64Slice, ps []float64) []float64 {
	scores := make([]float64, len(ps))
	size := len(values)
	if size > 0 {
		sort.Sort(values)
		for i, p := range ps {
			pos := p * float64(size+1)
			if pos < 1.0 {
				scores[i] = float64(values[0])
			} else if pos >= float64(size) {
				scores[i] = float64(values[size-1])
			} else {
				lower := float64(values[int(pos)-1])
				upper := float64(values[int(pos)])
				scores[i] = lower + (pos-math.Floor(pos))*(upper-lower)
			}
		}
	}
	return scores
}

// SampleFloat64Snapshot is a read-only copy of another SampleFloat64.
type SampleFloat64Snapshot struct {
	count  int64
	values []float64
}

func NewSampleFloat64Snapshot(count int64, values []float64) *SampleFloat64Snapshot {
	return &SampleFloat64Snapshot{
		count:  count,
		values: values,
	}
}

// Clear panics.
func (*SampleFloat64Snapshot) Clear() {
	panic("Clear called on a SampleFloat64Snapshot")
}

// Count returns the count of inputs at the time the snapshot was taken.
func (s *SampleFloat64Snapshot) Count() int64 { return s.count }

// Max returns the maximal value at the time the snapshot was taken.
func (s *SampleFloat64Snapshot) Max() float64 { return SampleFloat64Max(s.values) }

// Mean returns the mean value at the time the snapshot was taken.
func (s *SampleFloat64Snapshot) Mean() float64 { return SampleFloat64Mean(s.values) }

// Min returns the minimal value at the time the snapshot was taken.
func (s *SampleFloat64Snapshot) Min() float64 { return SampleFloat64Min(s.values) }

// Percentile returns an arbitrary percentile of values at the time the
// snapshot was taken.
func (s *SampleFloat64Snapshot) Percentile(p float64) float64 {
	return SampleFloat64Percentile(s.values, p)
}

// Percentiles returns a slice of arbitrary percentiles of values at the time
// the snapshot was taken.
func (s *SampleFloat64Snapshot) Percentiles(ps []float64) []float64 {
	return SampleFloat64Percentiles(s.values, ps)
}

// Size returns the size of the SampleFloat64 at the time the snapshot was taken.
func (s *SampleFloat64Snapshot) Size() int { return len(s.values) }

// Snapshot returns the snapshot.
func (s *SampleFloat64Snapshot) Snapshot() SampleFloat64 { return s }

// StdDev returns the standard deviation of values at the time the snapshot was
// taken.
func (s *SampleFloat64Snapshot) StdDev() float64 { return SampleFloat64StdDev(s.values) }

// Sum returns the sum of values at the time the snapshot was taken.
func (s *SampleFloat64Snapshot) Sum() float64 { return SampleFloat64Sum(s.values) }

// Update panics.
func (*SampleFloat64Snapshot) Update(float64) {
	panic("Update called on a SampleFloat64Snapshot")
}

// Values returns a copy of the values in the SampleFloat64.
func (s *SampleFloat64Snapshot) Values() []float64 {
	values := make([]float64, len(s.values))
	copy(values, s.values)
	return values
}

// Variance returns the variance of values at the time the snapshot was taken.
func (s *SampleFloat64Snapshot) Variance() float64 { return SampleFloat64Variance(s.values) }

// SampleFloat64StdDev returns the standard deviation of the slice of float64.
func SampleFloat64StdDev(values []float64) float64 {
	return math.Sqrt(SampleFloat64Variance(values))
}

// SampleFloat64Sum returns the sum of the slice of float64.
func SampleFloat64Sum(values []float64) float64 {
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum
}

// SampleFloat64Variance returns the variance of the slice of float64.
func SampleFloat64Variance(values []float64) float64 {
	if 0 == len(values) {
		return 0.0
	}
	m := SampleFloat64Mean(values)
	var sum float64
	for _, v := range values {
		d := float64(v) - m
		sum += d * d
	}
	return sum / float64(len(values))
}

// A uniform SampleFloat64 using Vitter's Algorithm R.
//
// <http://www.cs.umd.edu/~samir/498/vitter.pdf>
type UniformSampleFloat64 struct {
	count         int64
	mutex         sync.Mutex
	reservoirSize int
	values        []float64
}

// NewUniformSampleFloat64 constructs a new uniform SampleFloat64 with the given reservoir
// size.
func NewUniformSampleFloat64(reservoirSize int) SampleFloat64 {
	if UseNilMetrics {
		return NilSampleFloat64{}
	}
	return &UniformSampleFloat64{
		reservoirSize: reservoirSize,
		values:        make([]float64, 0, reservoirSize),
	}
}

// Clear clears all SampleFloat64s.
func (s *UniformSampleFloat64) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.count = 0
	s.values = make([]float64, 0, s.reservoirSize)
}

// Count returns the number of SampleFloat64s recorded, which may exceed the
// reservoir size.
func (s *UniformSampleFloat64) Count() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.count
}

// Max returns the maximum value in the SampleFloat64, which may not be the maximum
// value ever to be part of the SampleFloat64.
func (s *UniformSampleFloat64) Max() float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleFloat64Max(s.values)
}

// Mean returns the mean of the values in the SampleFloat64.
func (s *UniformSampleFloat64) Mean() float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleFloat64Mean(s.values)
}

// Min returns the minimum value in the SampleFloat64, which may not be the minimum
// value ever to be part of the SampleFloat64.
func (s *UniformSampleFloat64) Min() float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleFloat64Min(s.values)
}

// Percentile returns an arbitrary percentile of values in the SampleFloat64.
func (s *UniformSampleFloat64) Percentile(p float64) float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleFloat64Percentile(s.values, p)
}

// Percentiles returns a slice of arbitrary percentiles of values in the
// SampleFloat64.
func (s *UniformSampleFloat64) Percentiles(ps []float64) []float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleFloat64Percentiles(s.values, ps)
}

// Size returns the size of the SampleFloat64, which is at most the reservoir size.
func (s *UniformSampleFloat64) Size() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return len(s.values)
}

// Snapshot returns a read-only copy of the SampleFloat64.
func (s *UniformSampleFloat64) Snapshot() SampleFloat64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	values := make([]float64, len(s.values))
	copy(values, s.values)
	return &SampleFloat64Snapshot{
		count:  s.count,
		values: values,
	}
}

// StdDev returns the standard deviation of the values in the SampleFloat64.
func (s *UniformSampleFloat64) StdDev() float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleFloat64StdDev(s.values)
}

// Sum returns the sum of the values in the SampleFloat64.
func (s *UniformSampleFloat64) Sum() float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleFloat64Sum(s.values)
}

// Update SampleFloat64s a new value.
func (s *UniformSampleFloat64) Update(v float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.count++
	if len(s.values) < s.reservoirSize {
		s.values = append(s.values, v)
	} else {
		r := rand.Int63n(s.count)
		if r < int64(len(s.values)) {
			s.values[int(r)] = v
		}
	}
}

// Values returns a copy of the values in the SampleFloat64.
func (s *UniformSampleFloat64) Values() []float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	values := make([]float64, len(s.values))
	copy(values, s.values)
	return values
}

// Variance returns the variance of the values in the SampleFloat64.
func (s *UniformSampleFloat64) Variance() float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SampleFloat64Variance(s.values)
}

// expDecaySampleFloat64 represents an individual SampleFloat64 in a heap.
type expDecaySampleFloat64 struct {
	k float64
	v float64
}

func newExpDecaySampleFloat64Heap(reservoirSize int) *expDecaySampleFloat64Heap {
	return &expDecaySampleFloat64Heap{make([]expDecaySampleFloat64, 0, reservoirSize)}
}

// expDecaySampleFloat64Heap is a min-heap of expDecaySampleFloat64s.
// The internal implementation is copied from the standard library's container/heap
type expDecaySampleFloat64Heap struct {
	s []expDecaySampleFloat64
}

func (h *expDecaySampleFloat64Heap) Clear() {
	h.s = h.s[:0]
}

func (h *expDecaySampleFloat64Heap) Push(s expDecaySampleFloat64) {
	n := len(h.s)
	h.s = h.s[0 : n+1]
	h.s[n] = s
	h.up(n)
}

func (h *expDecaySampleFloat64Heap) Pop() expDecaySampleFloat64 {
	n := len(h.s) - 1
	h.s[0], h.s[n] = h.s[n], h.s[0]
	h.down(0, n)

	n = len(h.s)
	s := h.s[n-1]
	h.s = h.s[0 : n-1]
	return s
}

func (h *expDecaySampleFloat64Heap) Size() int {
	return len(h.s)
}

func (h *expDecaySampleFloat64Heap) Values() []expDecaySampleFloat64 {
	return h.s
}

func (h *expDecaySampleFloat64Heap) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !(h.s[j].k < h.s[i].k) {
			break
		}
		h.s[i], h.s[j] = h.s[j], h.s[i]
		j = i
	}
}

func (h *expDecaySampleFloat64Heap) down(i, n int) {
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && !(h.s[j1].k < h.s[j2].k) {
			j = j2 // = 2*i + 2  // right child
		}
		if !(h.s[j].k < h.s[i].k) {
			break
		}
		h.s[i], h.s[j] = h.s[j], h.s[i]
		i = j
	}
}

type float64Slice []float64

func (p float64Slice) Len() int           { return len(p) }
func (p float64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p float64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
