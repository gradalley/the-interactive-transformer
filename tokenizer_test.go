package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenizer(t *testing.T) {
	tokenizer, err := NewTokenizer("/Users/joshcarp/Documents/the-interactive-transformer/gpt2_tokenizer.bin")
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	encoded, err := tokenizer.Encode("hello there")
	fmt.Println("encoded: ", encoded)
	decoded, err := tokenizer.Decode(encoded)
	fmt.Println("decoded: ", decoded)
	require.Equal(t, decoded, "hello there")
}

func TestTokenizerJson(t *testing.T) {
	tokenizer, err := NewTokenizerJson("/Users/joshcarp/Documents/the-interactive-transformer/tokenizer.json")
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	encoded, err := tokenizer.Encode("hello there")
	fmt.Println("encoded: ", encoded)
	decoded, err := tokenizer.Decode(encoded)
	fmt.Println("decoded: ", decoded)
	require.Equal(t, decoded, "hello there")
}
