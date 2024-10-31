package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	assertI := assert.New(t)
	assertI.Equal(GetEnv("TestGin", "test"), "test", "they should be equal")
	_ = os.Setenv("TestGin", "stage")
	assertI.NotEqual(GetEnv("TestGin", "test"), "test", "they should be not equal")
}
