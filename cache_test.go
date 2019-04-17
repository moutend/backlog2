package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashQuery(t *testing.T) {
	q1 := url.Values{}
	q1 = nil
	assert.Equal(t, "", hashQuery(q1))

	q2 := url.Values{}
	q2.Add("foo", "bar")
	q2.Add("fizz", "buzz")
	q2.Add("hello", "world")
	h2 := hashQuery(q2)
	assert.NotEqual(t, "", h2)

	q3 := url.Values{}
	q3.Add("foo", "bar")
	q3.Add("fizz", "buzz")
	q3.Add("hello", "world")
	h3 := hashQuery(q3)
	assert.Equal(t, h2, h3)
}
