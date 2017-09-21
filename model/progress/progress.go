package progress

import (
	"fmt"
	"math"
	"os"
	"sync"
	"sync/atomic"
)

// ReportFunc ...
type ReportFunc func(total, progress uint64, pct float64)

// Component ...
type Component struct {
	// status is the parent status
	status *Status

	total, progress *uint64
	// ready is a boolean disguised as an int32, so atomic operations are
	// possible.
	ready *int32
}

// Status ...
type Status struct {
	reportFunc ReportFunc
	mu         sync.RWMutex
	components []*Component
}

// New ...
func New(reporter ReportFunc) *Status {
	return &Status{reportFunc: reporter}
}

// NewComponent returns a new sub-component of the aggregate status.
func (s *Status) NewComponent() *Component {
	c := &Component{
		total:    new(uint64),
		progress: new(uint64),
		ready:    new(int32),
		status:   s,
	}
	s.mu.Lock()
	s.components = append(s.components, c)
	s.mu.Unlock()
	return c
}

func (s *Status) update() {
	var total, progress uint64
	ready := true
	s.mu.RLock()
	for _, component := range s.components {
		total += atomic.LoadUint64(component.total)
		progress += atomic.LoadUint64(component.progress)
		if atomic.LoadInt32(component.ready) != 1 {
			ready = false
		}
	}
	s.mu.RUnlock()
	pct := math.NaN()
	if ready {
		pct = float64(progress) / float64(total) * 100
	}
	s.reportFunc(total, progress, pct)
}

// Total sets the total and sets ready to true.
func (c *Component) Total(total uint64) {
	atomic.StoreUint64(c.total, total)
	atomic.StoreInt32(c.ready, 1)
	c.status.update()
}

// Progress sets the progress.
func (c *Component) Progress(progress uint64) {
	atomic.StoreUint64(c.progress, progress)
	total := atomic.LoadUint64(c.total)
	if progress > total {
		fmt.Fprintf(os.Stderr, "Progress(): progress (%d) > total (%d)\n", progress, total)
	}
	c.status.update()
}

// Increment increments the progress by the specified amount.
func (c *Component) Increment(inc uint64) {
	prog := atomic.AddUint64(c.progress, inc)
	total := atomic.LoadUint64(c.total)
	if prog > total {
		fmt.Fprintf(os.Stderr, "Increment(): progress (%d) > total (%d)\n", prog, total)
	}
	c.status.update()
}

func (c *Component) Reporter() ReportFunc {
	return func(total, progress uint64, _ float64) {
		atomic.StoreUint64(c.total, total)
		atomic.StoreUint64(c.progress, progress)
		c.status.update()
	}
}
