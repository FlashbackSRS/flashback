package progress

import (
	"math"
	"testing"

	"github.com/flimzy/diff"
)

type result struct {
	total, progress uint64
	percent         float64
}

func TestProgress(t *testing.T) {
	var results []result
	fn := func(total, prog uint64, pct float64) {
		results = append(results, result{
			total:    total,
			progress: prog,
			percent:  pct,
		})
	}
	status := New(fn)
	cmp1 := status.NewComponent()
	cmp2 := status.NewComponent()
	cmp1.Progress(1)
	cmp1.Progress(5)
	cmp2.Total(100)
	cmp1.Total(300)
	cmp1.Progress(25)
	cmp1.Progress(200)
	cmp2.Progress(100)
	cmp1.Progress(300)

	expected := []result{
		{0, 1, math.NaN()},
		{0, 5, math.NaN()},
		{100, 5, math.NaN()},
		{400, 5, 1.25},
		{400, 25, 6.25},
		{400, 200, 50},
		{400, 300, 75},
		{400, 400, 100},
	}

	if d := diff.Interface(expected, results); d != nil {
		t.Error(d)
	}
}
