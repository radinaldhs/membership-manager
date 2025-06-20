package config

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type basicConfig struct {
	Interface any     `env:"INTERFACE"`
	String    string  `env:"STRING"`
	Int       int     `env:"INT"`
	Uint      uint    `env:"UINT"`
	Float64   float64 `env:"FLOAT64"`
}

func testBasics(t *testing.T) {
	var iface interface{} = "interface"
	str := "string"
	i := 123
	ui := uint(456)
	f64 := 1.23

	if err := os.Setenv("INTERFACE", "interface"); err != nil {
		t.Fatal(err.Error())
	}

	if err := os.Setenv("STRING", "string"); err != nil {
		t.Fatal(err.Error())
	}

	if err := os.Setenv("INT", "123"); err != nil {
		t.Fatal(err.Error())
	}

	if err := os.Setenv("UINT", "456"); err != nil {
		t.Fatal(err.Error())
	}

	if err := os.Setenv("FLOAT64", "1.23"); err != nil {
		t.Fatal(err.Error())
	}

	var cfg basicConfig
	if err := LoadEnv(&cfg); err != nil {
		t.Error(err.Error())
		return
	}

	if assert.NotNil(t, cfg.Interface) {
		assert.Equal(t, iface, cfg.Interface)
	}

	assert.Equal(t, str, cfg.String)
	assert.Equal(t, i, cfg.Int)
	assert.Equal(t, ui, cfg.Uint)
	assert.Equal(t, f64, cfg.Float64)
}

type boolConfig struct {
	BoolStrTrue bool `env:"BOOL_STR"`
	BoolNumTrue bool `env:"BOOL_NUM"`
}

func testBool(t *testing.T) {
	boolStr := "true"
	boolNum := "1"

	if err := os.Setenv("BOOL_STR", boolStr); err != nil {
		t.Fatal(err.Error())
	}

	if err := os.Setenv("BOOL_NUM", boolNum); err != nil {
		t.Fatal(err.Error())
	}

	var cfg boolConfig
	if err := LoadEnv(&cfg); err != nil {
		t.Error(err.Error())
		return
	}

	assert.True(t, cfg.BoolStrTrue)
	assert.True(t, cfg.BoolNumTrue)
}

type memberConfig struct {
	Nested string `env:"NESTED"`
}

type nestedConfig struct {
	Member memberConfig
}

func testNested(t *testing.T) {
	enval := "nested"
	if err := os.Setenv("NESTED", enval); err != nil {
		t.Fatal(err.Error())
	}

	var cfg nestedConfig
	if err := LoadEnv(&cfg); err != nil {
		t.Error(err.Error())
		return
	}

	assert.Equal(t, enval, cfg.Member.Nested)
}

type withPointerConfig struct {
	StringPtr *string `env:"STRING_PTR"`
}

func testPointer(t *testing.T) {
	enval := "string pointer"
	if err := os.Setenv("STRING_PTR", enval); err != nil {
		t.Fatal(err.Error())
	}

	var cfg withPointerConfig
	if err := LoadEnv(&cfg); err != nil {
		t.Error(err.Error())
		return
	}

	if assert.NotNil(t, cfg.StringPtr) {
		assert.Equal(t, enval, *cfg.StringPtr)
	}
}

func TestLoadEnv(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name:     "Success on parsing basic values",
			testFunc: testBasics,
		},
		{
			name:     "Success on parsing boolean value variations",
			testFunc: testBool,
		},
		{
			name:     "Success on nested parsing",
			testFunc: testNested,
		},
		{
			name:     "Success on parsing to pointer",
			testFunc: testPointer,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func Test_parseTextUnmarshaler(t *testing.T) {
	type args struct {
		envVars    environmentVariables
		val        any
		envk       string
		defaultVal string
	}
	tests := []struct {
		name           string
		args           args
		expectResult   any
		wantOk         bool
		wantErr        bool
		wantResultFunc func(t *testing.T, expect, got any)
	}{
		{
			name: "Unmarshal text representation of time.Time",
			args: args{
				envVars: map[string]string{
					"TIME": "2025-03-06T02:46:47+00:00",
				},
				val:        &time.Time{},
				envk:       "TIME",
				defaultVal: "",
			},
			expectResult: func() time.Time {
				expect, _ := time.Parse(time.RFC3339, "2025-03-06T02:46:47+00:00")
				return expect
			}(),
			wantOk:  true,
			wantErr: false,
			wantResultFunc: func(t *testing.T, expect, got any) {
				expectT := expect.(time.Time)
				gotT := *got.(*time.Time)
				assert.Equal(t, expectT, gotT)
			},
		},
		{
			name: "Unmarshal invalid text representation of time.Time",
			args: args{
				envVars: map[string]string{
					"TIME": "I don't know which time is it",
				},
				val:        &time.Time{},
				envk:       "TIME",
				defaultVal: "",
			},
			expectResult: time.Time{},
			wantOk:       false,
			wantErr:      true,
			wantResultFunc: func(t *testing.T, expect, got any) {
				expectT := expect.(time.Time)
				gotT := *got.(*time.Time)
				assert.Equal(t, expectT, gotT)
			},
		},
		{
			name: "Unmarshal type that did not implement encoding.TextUnmarshaler",
			args: args{
				envVars: map[string]string{
					"STRING": "string",
				},
				val:        "",
				envk:       "STRING",
				defaultVal: "",
			},
			expectResult: "",
			wantOk:       false,
			wantErr:      false,
			wantResultFunc: func(t *testing.T, expect, got any) {
				assert.Equal(t, "", got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := parseTextUnmarshaler(
				tt.args.envVars,
				reflect.ValueOf(tt.args.val),
				tt.args.envk,
				tt.args.defaultVal,
			)

			assert.Equal(t, tt.wantOk, ok)
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if !tt.wantErr {
				assert.Nil(t, err)
			}

			tt.wantResultFunc(t, tt.expectResult, tt.args.val)
		})
	}
}
