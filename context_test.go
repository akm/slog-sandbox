package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

type aKeyType struct{}
type bKeyType struct{}

var aKey aKeyType = aKeyType{}
var bKey bKeyType = bKeyType{}

type aWrapper struct {
	slog.Handler
}

var _ slog.Handler = (*aWrapper)(nil)

func newAWrapper(h slog.Handler) *aWrapper {
	return &aWrapper{Handler: h}
}

func (w *aWrapper) Handle(ctx context.Context, r slog.Record) error {
	v, ok := ctx.Value(aKey).(string)
	if ok {
		fmt.Printf("a: %s\n", v)
		r.AddAttrs(slog.String("a", v))
	}
	return w.Handler.Handle(ctx, r)
}

type bWrapper struct {
	slog.Handler
}

var _ slog.Handler = (*aWrapper)(nil)

func newBWrapper(h slog.Handler) *bWrapper {
	return &bWrapper{Handler: h}
}

func (w *bWrapper) Handle(ctx context.Context, r slog.Record) error {
	v, ok := ctx.Value(bKey).(int)
	if ok {
		r.AddAttrs(slog.Int("b", v))
	}
	return w.Handler.Handle(ctx, r)
}

type errorWrapper struct {
	slog.Handler
}

var _ slog.Handler = (*errorWrapper)(nil)

func newErrorWrapper(h slog.Handler) *errorWrapper {
	return &errorWrapper{Handler: h}
}

func (w *errorWrapper) Handle(ctx context.Context, r slog.Record) error {
	v, ok := ctx.Value(bKey).(int)
	if ok {
		r.AddAttrs(slog.Int("b", v))
	}
	return w.Handler.Handle(ctx, r)
}

type Password string

func (Password) LogValue() slog.Value {
	return slog.StringValue("********")
}

var _ slog.LogValuer = Password("")

func TestSlogWithContext(t *testing.T) {
	baseContext := context.Background()
	aContext := context.WithValue(baseContext, aKey, "a")
	abContext := context.WithValue(aContext, bKey, 2)

	nullWrapper := func(h slog.Handler) slog.Handler { return h }
	aWrapper := func(h slog.Handler) slog.Handler { return newAWrapper(h) }
	abWrapper := func(h slog.Handler) slog.Handler { return newBWrapper(newAWrapper(h)) }

	patterns := []struct {
		name     string
		ctx      context.Context
		wrapper  func(slog.Handler) slog.Handler
		args     []any
		expected map[string]any
	}{
		{
			name:     "base with nullWrapper",
			ctx:      baseContext,
			wrapper:  nullWrapper,
			expected: map[string]any{"msg": "msg"},
		},
		{
			name:     "base with nullWrapper, with password",
			ctx:      baseContext,
			wrapper:  nullWrapper,
			args:     []any{"password", Password("123456")},
			expected: map[string]any{"msg": "msg", "password": "********"},
		},
		{
			name:     "base with nullWrapper and error",
			ctx:      baseContext,
			wrapper:  nullWrapper,
			args:     []any{"err", fmt.Errorf("test error")},
			expected: map[string]any{"msg": "msg", "err": "test error"},
		},
		{
			name:     "base with nullWrapper",
			ctx:      baseContext,
			wrapper:  aWrapper,
			expected: map[string]any{"msg": "msg"},
		},
		{
			name:     "base with nullWrapper",
			ctx:      baseContext,
			wrapper:  aWrapper,
			args:     []any{"a", "A"},
			expected: map[string]any{"msg": "msg", "a": "A"},
		},
		{
			name:     "aContext with nullWrapper",
			ctx:      aContext,
			wrapper:  nullWrapper,
			expected: map[string]any{"msg": "msg"},
		},
		{
			name:     "abContext with nullWrapper",
			ctx:      abContext,
			wrapper:  nullWrapper,
			expected: map[string]any{"msg": "msg"},
		},
		{
			name:     "aContext with aWrapper",
			ctx:      aContext,
			wrapper:  aWrapper,
			expected: map[string]any{"msg": "msg", "a": "a"},
		},
		{
			name:     "abContext with aWrapper",
			ctx:      abContext,
			wrapper:  aWrapper,
			expected: map[string]any{"msg": "msg", "a": "a"},
		},
		{
			name:     "abContext with abWrapper",
			ctx:      abContext,
			wrapper:  abWrapper,
			expected: map[string]any{"msg": "msg", "a": "a", "b": float64(2)},
		},
	}

	for _, ptn := range patterns {
		t.Run(ptn.name, func(t *testing.T) {
			buf := bytes.NewBufferString("")
			logger := slog.New(ptn.wrapper(slog.NewJSONHandler(buf, nil)))
			logger.InfoContext(ptn.ctx, "msg", ptn.args...)
			t.Logf("buf: %s", buf.String())
			d := map[string]any{}
			err := json.Unmarshal(buf.Bytes(), &d)
			assert.NoError(t, err)
			for k, v := range ptn.expected {
				assert.Equal(t, v, d[k])
			}
		})
	}
}
