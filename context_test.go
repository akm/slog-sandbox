package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlogWithContext(t *testing.T) {
	baseContext := context.Background()
	aContext := context.WithValue(baseContext, "a", "a")
	abContext := context.WithValue(aContext, "b", 2)

	patterns := []struct {
		name     string
		ctx      context.Context
		args     map[string]any
		expected map[string]any
	}{
		{
			name:     "base with msg only",
			ctx:      baseContext,
			expected: map[string]any{"msg": "msg"},
		},
		{
			name:     "aContext with msg only",
			ctx:      aContext,
			expected: map[string]any{"msg": "msg", "a": "a"},
		},
		{
			name:     "abContext with msg only",
			ctx:      abContext,
			expected: map[string]any{"msg": "msg", "a": "a", "b": 2},
		},
	}

	for _, ptn := range patterns {
		t.Run(ptn.name, func(t *testing.T) {
			buf := bytes.NewBufferString("")
			logger := slog.New(slog.NewJSONHandler(buf, nil))
			logger.InfoContext(ptn.ctx, "msg")
			d := map[string]any{}
			err := json.Unmarshal(buf.Bytes(), &d)
			assert.NoError(t, err)
		})
	}
}
