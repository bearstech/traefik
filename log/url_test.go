package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	_, e := Parse("syslog+udp://127.0.0.1:514/")
	assert.NoError(t, e)
}
