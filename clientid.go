package mauerspecht

import (
	"encoding/base64"
	"fmt"
)

const length = 24

type ClientId [length]byte

func (c *ClientId) Set(s string) error {
	if base64.URLEncoding.DecodedLen(len(s)) != length {
		return fmt.Errorf("wrong id string length (%d)", len(s))
	}
	var d [length]byte
	if n, err := base64.URLEncoding.Decode(d[:], []byte(s)); err != nil {
		return err
	} else if n != length {
		return fmt.Errorf("wrong length of decoded data (%d)", n)
	}
	copy(c[:], d[:])
	return nil
}

func (c ClientId) String() string {
	return base64.URLEncoding.EncodeToString(c[:])
}
