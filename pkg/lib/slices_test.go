package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunkSliceBy_1(t *testing.T) {
	s := []string{"1", "2", "3", "4", "5", "6", "7", "8"}

	chunks := ChunkSliceBy(s, 3)

	assert.Equal(t, [][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7", "8"}}, chunks)
}

func TestChunkSliceBy_2(t *testing.T) {
	s := []string{"1", "2", "3", "4", "5", "6", "7", "8"}

	chunks := ChunkSliceBy(s, 4)

	assert.Equal(t, [][]string{{"1", "2", "3", "4"}, {"5", "6", "7", "8"}}, chunks)
}
