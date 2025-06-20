package config

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHourDuration_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name         string
		hd           *HourDuration
		args         args
		wantDuration time.Duration
		wantErr      bool
	}{
		{
			name: "Valid hour duration",
			hd:   new(HourDuration),
			args: args{
				text: []byte("1"),
			},
			wantDuration: time.Duration(1) * time.Hour,
			wantErr:      false,
		},
		{
			name: "Invalid hour duration",
			hd:   new(HourDuration),
			args: args{
				text: []byte("abc"),
			},
			wantDuration: 0,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.hd.UnmarshalText(tt.args.text)

			assert.Equal(t, tt.wantDuration, time.Duration(*tt.hd))

			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if !tt.wantErr {
				assert.Nil(t, err)
			}
		})
	}
}

func TestMinuteDuration_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name         string
		md           *MinuteDuration
		args         args
		wantDuration time.Duration
		wantErr      bool
	}{
		{
			name: "Valid minute duration",
			md:   new(MinuteDuration),
			args: args{
				text: []byte("1"),
			},
			wantDuration: time.Duration(1) * time.Minute,
			wantErr:      false,
		},
		{
			name: "Invalid minute duration",
			md:   new(MinuteDuration),
			args: args{
				text: []byte("abc"),
			},
			wantDuration: 0,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.md.UnmarshalText(tt.args.text)

			assert.Equal(t, tt.wantDuration, time.Duration(*tt.md))

			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if !tt.wantErr {
				assert.Nil(t, err)
			}
		})
	}
}

func TestSecondDuration_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name         string
		sd           *SecondDuration
		args         args
		wantDuration time.Duration
		wantErr      bool
	}{
		{
			name: "Valid second duration",
			sd:   new(SecondDuration),
			args: args{
				text: []byte("1"),
			},
			wantDuration: time.Duration(1) * time.Second,
			wantErr:      false,
		},
		{
			name: "Invalid second duration",
			sd:   new(SecondDuration),
			args: args{
				text: []byte("abc"),
			},
			wantDuration: 0,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sd.UnmarshalText(tt.args.text)

			assert.Equal(t, tt.wantDuration, time.Duration(*tt.sd))

			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if !tt.wantErr {
				assert.Nil(t, err)
			}
		})
	}
}

func TestMillisecondDuration_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name         string
		msd          *MillisecondDuration
		args         args
		wantDuration time.Duration
		wantErr      bool
	}{
		{
			name: "Valid millisecond duration",
			msd:  new(MillisecondDuration),
			args: args{
				text: []byte("1"),
			},
			wantDuration: time.Duration(1) * time.Millisecond,
			wantErr:      false,
		},
		{
			name: "Invalid millisecond duration",
			msd:  new(MillisecondDuration),
			args: args{
				text: []byte("abc"),
			},
			wantDuration: 0,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msd.UnmarshalText(tt.args.text)

			assert.Equal(t, tt.wantDuration, time.Duration(*tt.msd))

			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if !tt.wantErr {
				assert.Nil(t, err)
			}
		})
	}
}

func encodeBase64(data []byte) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)

	return dst
}

func TestRawBase64Encoded_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name       string
		b64        *RawBase64Encoded
		args       args
		wantResult RawBase64Encoded
		wantErr    bool
	}{
		{
			name: "Valid base64-encoded data",
			b64:  new(RawBase64Encoded),
			args: args{
				text: encodeBase64([]byte("Hello World!")),
			},
			wantResult: RawBase64Encoded{
				Raw:     encodeBase64([]byte("Hello World!")),
				Decoded: []byte("Hello World!"),
			},
			wantErr: false,
		},
		{
			name: "Invalid base64-encoded data",
			b64:  new(RawBase64Encoded),
			args: args{
				text: []byte("invalid base64"),
			},
			wantResult: RawBase64Encoded{},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.b64.UnmarshalText(tt.args.text)
			assert.Equal(t, tt.wantResult, *tt.b64)

			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if !tt.wantErr {
				assert.Nil(t, err)
			}
		})
	}
}
