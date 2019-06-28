package crypto

import (
	"golang.org/x/crypto/nacl/box"

	"crypto/rand"
	"fmt"
	"io"
)

type Context struct {
	PubKey  *[32]byte
	privKey *[32]byte
}

func NewContext() (*Context, error) {
	var ctx = &Context{}
	var err error
	if ctx.PubKey, ctx.privKey, err = box.GenerateKey(rand.Reader); err != nil {
		return nil, err
	}
	return ctx, nil
}

func (ctx *Context) Encrypt(peerPubKey *[32]byte, cleartext []byte) (ciphertext []byte, err error) {
	var nonce [24]byte
	if _, err = io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return
	}
	ciphertext = box.Seal(nonce[:], cleartext, &nonce, peerPubKey, ctx.privKey)
	return
}

func (ctx *Context) Decrypt(peerPubKey *[32]byte, ciphertext []byte) (cleartext []byte, err error) {
	if l := len(ciphertext); l < 24 {
		return nil, fmt.Errorf("ciphertext too short (%d)", l)
	}
	var nonce [24]byte
	copy(nonce[:], ciphertext[:24])
	var ok bool
	cleartext, ok = box.Open(nil, ciphertext[24:], &nonce, peerPubKey, ctx.privKey)
	if !ok {
		return nil, fmt.Errorf("decryption error")
	}
	return
}
