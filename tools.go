// +build tools

package main

import (
	_ "github.com/FiloSottile/mkcert"
	_ "golang.org/x/lint/golint"
	_ "mvdan.cc/gofumpt/gofumports"
)
