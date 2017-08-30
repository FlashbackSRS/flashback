package test

import (
	"testing"

	_ "github.com/flimzy/kivik/driver/memory"
)

func init() {
	RegisterMemoryDBSuite()
}

func TestMemory(t *testing.T) {
	MemoryTest(t)
}
