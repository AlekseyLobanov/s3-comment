package main

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown"
)

const HASH_LEN = 12 // taken from isso

func CalculateUserHash(email string, secret string) string {
	hash := sha1.New()
	hash.Write([]byte(strings.ToLower(email)))
	hash.Write([]byte(secret))
	return fmt.Sprintf("%x", hash.Sum(nil))[:HASH_LEN]
}

func RenderMarkdown(input_md string) string {
	md := []byte(input_md)
	output := markdown.ToHTML(md, nil, nil)
	return string(output)
}
