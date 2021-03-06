package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderConfigs_Copy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		a    *ProviderConfigs
	}{
		{
			"nil",
			nil,
		},
		{
			"empty",
			&ProviderConfigs{},
		},
		{
			"same_enabled",
			&ProviderConfigs{
				{
					"null": map[string]interface{}{
						"attr":  "value",
						"count": 10,
					},
				},
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			r := tc.a.Copy()
			assert.Equal(t, tc.a, r)
		})
	}
}

func TestProviderConfigs_Merge(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		a    *ProviderConfigs
		b    *ProviderConfigs
		r    *ProviderConfigs
	}{
		{
			"nil_a",
			nil,
			&ProviderConfigs{},
			&ProviderConfigs{},
		},
		{
			"nil_b",
			&ProviderConfigs{},
			nil,
			&ProviderConfigs{},
		},
		{
			"nil_both",
			nil,
			nil,
			nil,
		},
		{
			"empty",
			&ProviderConfigs{},
			&ProviderConfigs{},
			&ProviderConfigs{},
		},
		{
			"appends",
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr":  "n",
					"count": 10,
				},
			}},
			&ProviderConfigs{{
				"template": map[string]interface{}{
					"attr":  "t",
					"count": 5,
				},
			}},
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr":  "n",
					"count": 10,
				},
			}, {
				"template": map[string]interface{}{
					"attr":  "t",
					"count": 5,
				},
			}},
		},
		{
			"empty_one",
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr":  "n",
					"count": 10,
				},
			}},
			&ProviderConfigs{},
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr":  "n",
					"count": 10,
				},
			}},
		},
		{
			"empty_two",
			&ProviderConfigs{},
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr":  "n",
					"count": 10,
				},
			}},
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr":  "n",
					"count": 10,
				},
			}},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			r := tc.a.Merge(tc.b)
			assert.Equal(t, tc.r, r)
		})
	}
}

func TestProviderConfigs_Finalize(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		i    *ProviderConfigs
		r    *ProviderConfigs
	}{
		{
			"empty",
			&ProviderConfigs{},
			&ProviderConfigs{},
		},
		{
			"with_name",
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr":  "n",
					"count": 10,
				},
			}},
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr":  "n",
					"count": 10,
				},
			}},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			tc.i.Finalize()
			assert.Equal(t, tc.r, tc.i)
		})
	}
}

func TestProviderConfigs_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		i       *ProviderConfigs
		isValid bool
	}{
		{
			"nil",
			nil,
			false,
		},
		{
			"empty",
			&ProviderConfigs{},
			true,
		},
		{
			"valid",
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr":  "n",
					"count": 10,
				},
			}},
			true,
		},
		{
			"empty provider map",
			&ProviderConfigs{{}},
			false,
		}, {
			"multiple",
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr":  "n",
					"count": 10,
				},
			}, {
				"template": map[string]interface{}{
					"foo": "bar",
				},
			}},
			true,
		}, {
			"alias",
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr":  "n",
					"count": 10,
				},
			}, {
				"null": map[string]interface{}{
					"alias": "negative",
					"attr":  "abc",
					"count": -2,
				},
			}},
			true,
		}, {
			"duplicate",
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr":  "n",
					"count": 10,
				},
			}, {
				"null": map[string]interface{}{
					"attr":  "abc",
					"count": -2,
				},
			}},
			false,
		}, {
			"duplicate alias",
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"alias": "alias",
					"attr":  "n",
					"count": 10,
				},
			}, {
				"null": map[string]interface{}{
					"alias": "alias",
					"attr":  "abc",
					"count": -2,
				},
			}},
			false,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			err := tc.i.Validate()
			if tc.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestProviderConfigs_GoString(t *testing.T) {
	t.Parallel()

	// only testing provider cases with one argument since map order is random
	cases := []struct {
		name     string
		i        *ProviderConfigs
		expected string
	}{
		{
			"nil",
			nil,
			`(*ProviderConfigs)(nil)`,
		},
		{
			"empty",
			&ProviderConfigs{},
			`{}`,
		},
		{
			"single config",
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"count": "10",
				},
			}},
			fmt.Sprintf(`{&map[null:%s]}`, redactMessage),
		},
		{
			"multiple configs, same provider",
			&ProviderConfigs{{
				"null": map[string]interface{}{
					"attr": "n",
				},
			}, {
				"null": map[string]interface{}{
					"alias": "negative",
				},
			}},
			fmt.Sprintf(`{&map[null:%s], &map[null:%s]}`,
				redactMessage, redactMessage),
		},
		{
			"multiple configs, different provider",
			&ProviderConfigs{{
				"firewall": map[string]interface{}{
					"hostname": "127.0.0.10",
					"username": "username",
					"password": "password123",
				},
			}, {
				"loadbalancer": map[string]interface{}{
					"address": "10.10.10.10",
					"api_key": "abcd123",
				},
			}},
			fmt.Sprintf(`{&map[firewall:%s], &map[loadbalancer:%s]}`,
				redactMessage, redactMessage),
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			actual := tc.i.GoString()
			assert.Equal(t, tc.expected, actual)
		})
	}
}
