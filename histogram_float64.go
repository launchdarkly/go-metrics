package metrics

import "sync"

// Histograms calculate distribution statistics from a series of float64 values.
type HistogramFloat64 interface {
	Clear() HistogramFloat64 // atomically clears and returns a snapshot
	Count() int64
	Max() float64
	Mean() float64
	Min() float64
	Percentile(float64) float64
	Percentiles([]float64) []float64
	Sample() SampleFloat64
	Snapshot() HistogramFloat64
	StdDev() float64
	Sum() float64
	Update(float64)
	Variance() float64
}

// GetOrRegisterHistogram returns an existing Histogram or constructs and
// registers a new StandardHistogramFloat64.
func GetOrRegisterHistogramFloat64(name string, r Registry, s SampleFloat64) HistogramFloat64 {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() HistogramFloat64 { return NewHistogramFloat64(s) }).(HistogramFloat64)
}

// NewHistogram constructs a new StandardHistogramFloat64 from a Sample.
func NewHistogramFloat64(s SampleFloat64) HistogramFloat64 {
	if UseNilMetrics {
		return NilHistogramFloat64{}
	}
	return &StandardHistogramFloat64{sample: s}
}

// NewRegisteredHistogram constructs and registers a new StandardHistogramFloat64 from
// a Sample.
func NewRegisteredHistogramFloat64(name string, r Registry, s SampleFloat64) HistogramFloat64 {
	c := NewHistogramFloat64(s)
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// HistogramSnapshotFloat64 is a read-only copy of another Histogram.
type HistogramSnapshotFloat64 struct {
	sample *SampleFloat64Snapshot
}

// Clear panics.
func (*HistogramSnapshotFloat64) Clear() HistogramFloat64 {
	panic("Clear called on a HistogramSnapshotFloat64")
}

// Count returns the number of samples recorded at the time the snapshot was
// taken.
func (h *HistogramSnapshotFloat64) Count() int64 { return h.sample.Count() }

// Max returns the maximum value in the sample at the time the snapshot was
// taken.
func (h *HistogramSnapshotFloat64) Max() float64 { return h.sample.Max() }

// Mean returns the mean of the values in the sample at the time the snapshot
// was taken.
func (h *HistogramSnapshotFloat64) Mean() float64 { return h.sample.Mean() }

// Min returns the minimum value in the sample at the time the snapshot was
// taken.
func (h *HistogramSnapshotFloat64) Min() float64 { return h.sample.Min() }

// Percentile returns an arbitrary percentile of values in the sample at the
// time the snapshot was taken.
func (h *HistogramSnapshotFloat64) Percentile(p float64) float64 {
	return h.sample.Percentile(p)
}

// Percentiles returns a slice of arbitrary percentiles of values in the sample
// at the time the snapshot was taken.
func (h *HistogramSnapshotFloat64) Percentiles(ps []float64) []float64 {
	return h.sample.Percentiles(ps)
}

// Sample returns the Sample underlying the histogram.
func (h *HistogramSnapshotFloat64) Sample() SampleFloat64 { return h.sample }

// Snapshot returns the snapshot.
func (h *HistogramSnapshotFloat64) Snapshot() HistogramFloat64 { return h }

// StdDev returns the standard deviation of the values in the sample at the
// time the snapshot was taken.
func (h *HistogramSnapshotFloat64) StdDev() float64 { return h.sample.StdDev() }

// Sum returns the sum in the sample at the time the snapshot was taken.
func (h *HistogramSnapshotFloat64) Sum() float64 { return h.sample.Sum() }

// Update panics.
func (*HistogramSnapshotFloat64) Update(float64) {
	panic("Update called on a HistogramSnapshotFloat64")
}

// Variance returns the variance of inputs at the time the snapshot was taken.
func (h *HistogramSnapshotFloat64) Variance() float64 { return h.sample.Variance() }

// NilHistogramFloat64 is a no-op Histogram.
type NilHistogramFloat64 struct{}

// Clear is a no-op.
func (NilHistogramFloat64) Clear() HistogramFloat64 { return NilHistogramFloat64{} }

// Count is a no-op.
func (NilHistogramFloat64) Count() int64 { return 0 }

// Max is a no-op.
func (NilHistogramFloat64) Max() float64 { return 0 }

// Mean is a no-op.
func (NilHistogramFloat64) Mean() float64 { return 0.0 }

// Min is a no-op.
func (NilHistogramFloat64) Min() float64 { return 0 }

// Percentile is a no-op.
func (NilHistogramFloat64) Percentile(p float64) float64 { return 0.0 }

// Percentiles is a no-op.
func (NilHistogramFloat64) Percentiles(ps []float64) []float64 {
	return make([]float64, len(ps))
}

// Sample is a no-op.
func (NilHistogramFloat64) Sample() SampleFloat64 { return NilSampleFloat64{} }

// Snapshot is a no-op.
func (NilHistogramFloat64) Snapshot() HistogramFloat64 { return NilHistogramFloat64{} }

// StdDev is a no-op.
func (NilHistogramFloat64) StdDev() float64 { return 0.0 }

// Sum is a no-op.
func (NilHistogramFloat64) Sum() float64 { return 0 }

// Update is a no-op.
func (NilHistogramFloat64) Update(v float64) {}

// Variance is a no-op.
func (NilHistogramFloat64) Variance() float64 { return 0.0 }

// StandardHistogramFloat64 is the standard implementation of a Histogram and uses a
// Sample to bound its memory use.
type StandardHistogramFloat64 struct {
	sample SampleFloat64
	mutex  sync.Mutex
}

// Clear clears the histogram and its sample.
func (h *StandardHistogramFloat64) Clear() HistogramFloat64 {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	hSnap := &HistogramSnapshotFloat64{sample: h.sample.Snapshot().(*SampleFloat64Snapshot)}
	h.sample.Clear()
	return hSnap
}

// Count returns the number of samples recorded since the histogram was last
// cleared.
func (h *StandardHistogramFloat64) Count() int64 { return h.sample.Count() }

// Max returns the maximum value in the sample.
func (h *StandardHistogramFloat64) Max() float64 { return h.sample.Max() }

// Mean returns the mean of the values in the sample.
func (h *StandardHistogramFloat64) Mean() float64 { return h.sample.Mean() }

// Min returns the minimum value in the sample.
func (h *StandardHistogramFloat64) Min() float64 { return h.sample.Min() }

// Percentile returns an arbitrary percentile of the values in the sample.
func (h *StandardHistogramFloat64) Percentile(p float64) float64 {
	return h.sample.Percentile(p)
}

// Percentiles returns a slice of arbitrary percentiles of the values in the
// sample.
func (h *StandardHistogramFloat64) Percentiles(ps []float64) []float64 {
	return h.sample.Percentiles(ps)
}

// Sample returns the Sample underlying the histogram.
func (h *StandardHistogramFloat64) Sample() SampleFloat64 { return h.sample }

// Snapshot returns a read-only copy of the histogram.
func (h *StandardHistogramFloat64) Snapshot() HistogramFloat64 {
	return &HistogramSnapshotFloat64{sample: h.sample.Snapshot().(*SampleFloat64Snapshot)}
}

// StdDev returns the standard deviation of the values in the sample.
func (h *StandardHistogramFloat64) StdDev() float64 { return h.sample.StdDev() }

// Sum returns the sum in the sample.
func (h *StandardHistogramFloat64) Sum() float64 { return h.sample.Sum() }

// Update samples a new value.
func (h *StandardHistogramFloat64) Update(v float64) { h.sample.Update(v) }

// Variance returns the variance of the values in the sample.
func (h *StandardHistogramFloat64) Variance() float64 { return h.sample.Variance() }
