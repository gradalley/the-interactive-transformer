package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

type Tokenizer2 struct {
	encoder       map[string]int
	decoder       map[int]string
	bpeRanks      map[string]int
	bpeRegex      *regexp.Regexp
	specialTokens map[string]int
}

func NewTokenizer2(path string) (*Tokenizer2, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var tokenizer struct {
		Encoder       map[string]int `json:"encoder"`
		BpeRanks      map[string]int `json:"bpe_ranks"`
		SpecialTokens map[string]int `json:"special_tokens"`
	}

	err = json.Unmarshal(data, &tokenizer)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	decoder := make(map[int]string)
	for token, id := range tokenizer.Encoder {
		decoder[id] = token
	}

	bpeRegex := regexp.MustCompile(`'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+`)

	return &Tokenizer2{
		encoder:       tokenizer.Encoder,
		decoder:       decoder,
		bpeRanks:      tokenizer.BpeRanks,
		bpeRegex:      bpeRegex,
		specialTokens: tokenizer.SpecialTokens,
	}, nil
}

func (t *Tokenizer2) Encode(text string) ([]int, error) {
	tokens := t.bpeRegex.FindAllString(text, -1)
	var encoded []int

	for _, token := range tokens {
		subTokens := t.bpe(token)
		for _, subToken := range subTokens {
			if id, ok := t.encoder[subToken]; ok {
				encoded = append(encoded, id)
			} else {
				return nil, fmt.Errorf("unknown token: %s", subToken)
			}
		}
	}

	return encoded, nil
}

func (t *Tokenizer2) Decode(tokens []int) (string, error) {
	var decoded []string

	for _, token := range tokens {
		if word, ok := t.decoder[token]; ok {
			decoded = append(decoded, word)
		} else {
			return "", fmt.Errorf("unknown token ID: %d", token)
		}
	}

	text := strings.Join(decoded, "")
	text = strings.ReplaceAll(text, "Ġ", " ")
	return strings.TrimSpace(text), nil
}

func (t *Tokenizer2) bpe(token string) []string {
	if specialID, ok := t.specialTokens[token]; ok {
		return []string{t.decoder[specialID]}
	}

	chars := strings.Split(token, "")
	if len(chars) == 1 {
		return chars
	}

	pairs := getPairs(chars)

	for {
		if len(pairs) == 0 {
			break
		}

		minPair := ""
		minRank := int(^uint(0) >> 1)

		for _, pair := range pairs {
			if rank, ok := t.bpeRanks[pair]; ok {
				if rank < minRank {
					minPair = pair
					minRank = rank
				}
			}
		}

		if minPair == "" {
			break
		}

		parts := strings.Split(minPair, ",")
		newChars := make([]string, 0, len(chars))
		i := 0

		for i < len(chars) {
			if i+1 < len(chars) && chars[i] == parts[0] && chars[i+1] == parts[1] {
				newChars = append(newChars, parts[0]+parts[1])
				i += 2
			} else {
				newChars = append(newChars, chars[i])
				i++
			}
		}

		chars = newChars
		if len(chars) == 1 {
			break
		}

		pairs = getPairs(chars)
	}

	return chars
}

func getPairs(chars []string) []string {
	var pairs []string
	for i := 0; i < len(chars)-1; i++ {
		pairs = append(pairs, chars[i]+","+chars[i+1])
	}
	return pairs
}
