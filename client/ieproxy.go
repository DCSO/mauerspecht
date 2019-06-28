//+build windows

package client

import "github.com/mattn/go-ieproxy"

func init() {
	ieproxy.OverrideEnvWithStaticProxy()
}
