package config

import (
	"encoding/base64"
	"strconv"
	"time"
)

func createDurationFromText(dst *time.Duration, text []byte, multiplier time.Duration) error {
	i64, err := strconv.ParseInt(string(text), 10, 64)
	if err != nil {
		return err
	}

	d := time.Duration(i64) * multiplier
	*dst = d

	return nil
}

type HourDuration time.Duration

func (hd *HourDuration) UnmarshalText(text []byte) error {
	return createDurationFromText((*time.Duration)(hd), text, time.Hour)
}

type MinuteDuration time.Duration

func (md *MinuteDuration) UnmarshalText(text []byte) error {
	return createDurationFromText((*time.Duration)(md), text, time.Minute)
}

type SecondDuration time.Duration

func (sd *SecondDuration) UnmarshalText(text []byte) error {
	return createDurationFromText((*time.Duration)(sd), text, time.Second)
}

type MillisecondDuration time.Duration

func (msd *MillisecondDuration) UnmarshalText(text []byte) error {
	return createDurationFromText((*time.Duration)(msd), text, time.Millisecond)
}

type RawBase64Encoded struct {
	Raw     []byte
	Decoded []byte
}

func (b64 *RawBase64Encoded) UnmarshalText(text []byte) error {
	decodeLen := base64.RawStdEncoding.DecodedLen(len(text))
	decoded := make([]byte, decodeLen)
	_, err := base64.RawStdEncoding.Decode(decoded, text)
	if err != nil {
		return err
	}

	b64.Raw = text
	b64.Decoded = decoded

	return nil
}
