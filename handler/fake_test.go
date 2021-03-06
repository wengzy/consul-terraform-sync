package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFake(t *testing.T) {
	cases := []struct {
		name        string
		expectError bool
		config      map[string]interface{}
		expect      *Fake
	}{
		{
			"happy path",
			false,
			map[string]interface{}{
				"name": "1",
				"err":  true,
			},
			&Fake{name: "1", err: true},
		},
		{
			"missing configuration",
			true,
			map[string]interface{}{},
			nil,
		},
		{
			"happy path + extra config",
			false,
			map[string]interface{}{
				"name":  "1",
				"extra": "stuff",
				"count": 8,
			},
			&Fake{name: "1", err: false},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewFake(tc.config)
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, h)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, *tc.expect, *h)
		})
	}
}

func TestFakeDo(t *testing.T) {
	cases := []struct {
		name      string
		next      Handler
		expectErr bool
	}{
		{
			"happy path - with next handler",
			&Fake{},
			false,
		},
		{
			"happy path - no next handler",
			nil,
			false,
		},
		{
			"error",
			nil,
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &Fake{}
			if tc.next != nil {
				h.SetNext(tc.next)
			}
			if tc.expectErr {
				h.err = true
			}

			err := h.Do(nil)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestFakeSetNext(t *testing.T) {
	cases := []struct {
		name string
	}{
		{
			"happy path",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &Fake{}
			h.SetNext(&Fake{})
			assert.NotNil(t, h.next)
		})
	}
}
