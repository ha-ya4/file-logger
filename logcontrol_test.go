package filelogger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoOutputProd(t *testing.T) {
	assert.True(t, ModeProduction.noOutputProd(DEBUG))
	assert.True(t, ModeProduction.noOutputProd(INFO))
	assert.True(t, ModeProduction.noOutputProd(WARN))
	assert.False(t, ModeProduction.noOutputProd(ERROR))
}