package config

import "gabe565.com/utils/bytefmt"

type Bytes int64

func (b Bytes) String() string {
	return bytefmt.Encode(int64(b))
}

func (b *Bytes) Set(s string) error {
	val, err := bytefmt.Decode(s)
	if err != nil {
		return err
	}
	*b = Bytes(val)
	return nil
}

func (b Bytes) Type() string {
	return "string"
}
