package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateUserHash(t *testing.T) {
	assert.NotEqual(
		t,
		CalculateUserHash("demo@example.com", "secret1"),
		CalculateUserHash("demo@example.com", "secret2"),
	)
	assert.NotEqual(
		t,
		CalculateUserHash("demo@example.com", "secret1"),
		CalculateUserHash("demo2@example.com", "secret1"),
	)
	assert.Equal(
		t,
		CalculateUserHash("demo@example.com", "secret1"),
		CalculateUserHash("demO@examplE.com", "secret1"),
	)
	assert.NotEqual(
		t,
		strings.ToLower(CalculateUserHash("demo@example.com", "secret1")),
		CalculateUserHash("demo2@example.com", "secret1"),
	)
}
