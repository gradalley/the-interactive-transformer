package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"sort"
)

const GPT2_EOT int32 = 50256

type Tokenizer struct {
	vocabSize  uint32
	tokenTable []string         // tokenTable maps token id to string
	tokenMap   map[string]int32 // tokenMap maps token to id
	init       bool
}

func newTokenizer(vocab []string) Tokenizer {
	tokenizer := Tokenizer{
		vocabSize:  uint32(len(vocab)),
		tokenTable: vocab,
		tokenMap:   make(map[string]int32),
		init:       true,
	}
	for i, token := range vocab {
		tokenizer.tokenMap[token] = int32(i)
	}
	return tokenizer
}

func NewTokenizer(filename string) (Tokenizer, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Tokenizer{}, err
	}
	defer f.Close()
	header := make([]uint32, 256)
	if err := binary.Read(f, binary.LittleEndian, header); err != nil {
		return Tokenizer{}, err
	}
	if header[0] != 20240328 || header[1] != 1 {
		return Tokenizer{}, errors.New("incorrect header for tokenizer")
	}
	tok := Tokenizer{
		vocabSize:  header[2],
		tokenTable: make([]string, header[2]),
		tokenMap:   make(map[string]int32),
		init:       true,
	}
	var length byte
	for i := range tok.tokenTable {
		if err := binary.Read(f, binary.LittleEndian, &length); err != nil {
			return tok, err
		}
		if length <= 0 {
			return tok, errors.New("tokenizer failure")
		}
		tokenBytes := make([]byte, length)
		if err := binary.Read(f, binary.LittleEndian, tokenBytes); err != nil {
			return tok, err
		}
		tok.tokenTable[i] = string(tokenBytes)
		tok.tokenMap[tok.tokenTable[i]] = int32(i)
	}
	return tok, nil
}

type TokenizerJSON struct {
	Version string `json:"version"`
	Model   struct {
		Type          string            `json:"type"`
		Vocab         map[string]int    `json:"vocab"`
		MergesData    []string          `json:"merges,omitempty"`
		SpecialTokens map[string]string `json:"special_tokens"`
	} `json:"model"`
}

func NewTokenizerJson(filename string) (Tokenizer, error) {
	// Read the JSON file
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return Tokenizer{}, err
	}

	// Unmarshal JSON into our struct
	var tokenizerData TokenizerJSON
	if err := json.Unmarshal(fileContent, &tokenizerData); err != nil {
		return Tokenizer{}, err
	}

	// Create a new Tokenizer instance
	tok := Tokenizer{
		vocabSize:  uint32(len(tokenizerData.Model.Vocab)),
		tokenTable: make([]string, len(tokenizerData.Model.Vocab)),
		tokenMap:   make(map[string]int32),
		init:       true,
	}

	// Create a slice of token-id pairs for sorting
	var tokenIDPairs []struct {
		Token string
		ID    int
	}
	for token, id := range tokenizerData.Model.Vocab {
		// Convert the first two bytes to the 'Ġ' character if they match 0xC4 0xA0
		if len(token) >= 2 && token[0] == 0xC4 && token[1] == 0xA0 {
			token = " " + token[2:]
		}
		tokenIDPairs = append(tokenIDPairs, struct {
			Token string
			ID    int
		}{token, id})
	}

	// Sort the token-id pairs by ID
	sort.Slice(tokenIDPairs, func(i, j int) bool {
		return tokenIDPairs[i].ID < tokenIDPairs[j].ID
	})

	// Populate tokenTable and tokenMap
	for i, pair := range tokenIDPairs {
		tok.tokenTable[i] = pair.Token
		tok.tokenMap[pair.Token] = int32(i)
	}

	return tok, nil
}

func (t Tokenizer) Decode(tokens []int32) (string, error) {
	s := ""
	for _, token := range tokens {
		if token >= int32(len(t.tokenTable)) {
			return "", errors.New("not valid token")
		}
		if token != GPT2_EOT {
			s += t.tokenTable[token]
		}
	}
	return s, nil
}

func (t Tokenizer) Encode(text string) ([]int32, error) {
	tokens := []int32{}
	for len(text) > 0 {
		longestMatch := ""
		longestMatchToken := int32(GPT2_EOT)
		for i := len(text); i > 0; i-- {
			subStr := text[:i]
			if token, exists := t.tokenMap[subStr]; exists {
				longestMatch = subStr
				longestMatchToken = token
				break
			}
		}
		if longestMatch == "" {
			// If no match found, treat the first character as an unknown token
			tokens = append(tokens, GPT2_EOT)
			text = text[1:]
		} else {
			tokens = append(tokens, longestMatchToken)
			text = text[len(longestMatch):]
		}
	}
	return tokens, nil
}
