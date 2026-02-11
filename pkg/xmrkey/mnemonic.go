/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2026 Laurynas Četyrkinas <laurynas@digilol.net>
 *
 * MoneroPay is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * MoneroPay is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with MoneroPay.  If not, see <https://www.gnu.org/licenses/>.
 */

package xmrkey

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidMnemonicLength = errors.New("mnemonic must be 25 words")
	ErrInvalidChecksum       = errors.New("mnemonic checksum invalid")
	ErrWordNotFound          = errors.New("word not found in wordlist")
	ErrUnknownLanguage       = errors.New("unknown mnemonic language")
)

// wordToIndexMaps holds reverse lookup maps for each language.
var wordToIndexMaps map[string]map[string]int

func init() {
	wordToIndexMaps = make(map[string]map[string]int, len(Languages))
	for name, lang := range Languages {
		m := make(map[string]int, len(lang.Words))
		for i, word := range lang.Words {
			m[word] = i
		}
		wordToIndexMaps[name] = m
	}
}

// KeyToMnemonic converts a 32-byte private spend key to a 25-word mnemonic
// using the specified language.
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/mnemonics/electrum-words.cpp#L374-L429
func KeyToMnemonic(key [32]byte, language string) (string, error) {
	lang, ok := Languages[language]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrUnknownLanguage, language)
	}

	words := make([]string, 24)
	n := uint32(len(lang.Words))

	// Process 8 chunks of 4 bytes each
	for i := range 8 {
		// Read 4 bytes as little-endian uint32
		val := binary.LittleEndian.Uint32(key[i*4 : (i+1)*4])

		// Calculate 3 word indices
		w1 := val % n
		w2 := ((val / n) + w1) % n
		w3 := (((val / n) / n) + w2) % n

		words[i*3] = lang.Words[w1]
		words[i*3+1] = lang.Words[w2]
		words[i*3+2] = lang.Words[w3]
	}

	// Calculate checksum
	checksumWord := calculateChecksumWithPrefix(words, lang.PrefixLen)

	return strings.Join(append(words, checksumWord), " "), nil
}

// MnemonicToKey converts a 25-word mnemonic to a 32-byte private spend key.
// It automatically detects the language from the mnemonic words and returns it.
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/mnemonics/electrum-words.cpp#L255-L347
func MnemonicToKey(mnemonic string) (key [32]byte, language string, err error) {
	inputWords := strings.Fields(strings.ToLower(mnemonic))
	if len(inputWords) != 25 {
		return key, "", ErrInvalidMnemonicLength
	}

	// Detect language by finding which wordlist contains all words
	var lang *Language
	var wordToIndex map[string]int
	for name, l := range Languages {
		allMatch := true
		for _, word := range inputWords {
			if _, ok := wordToIndexMaps[name][word]; !ok {
				allMatch = false
				break
			}
		}
		if allMatch {
			lang = l
			language = name
			wordToIndex = wordToIndexMaps[name]
			break
		}
	}

	if lang == nil {
		return key, "", ErrWordNotFound
	}

	// Verify checksum
	expectedChecksum := calculateChecksumWithPrefix(inputWords[:24], lang.PrefixLen)
	if inputWords[24] != expectedChecksum {
		return key, "", ErrInvalidChecksum
	}

	// Convert words to key
	n := len(lang.Words)
	for i := range 8 {
		w1 := wordToIndex[inputWords[i*3]]
		w2 := wordToIndex[inputWords[i*3+1]]
		w3 := wordToIndex[inputWords[i*3+2]]

		// Reverse the encoding
		val := uint32(w1)
		val += uint32(n) * uint32((w2-w1+n)%n)
		val += uint32(n) * uint32(n) * uint32((w3-w2+n)%n)

		binary.LittleEndian.PutUint32(key[i*4:(i+1)*4], val)
	}

	return key, language, nil
}

// calculateChecksumWithPrefix computes the checksum word for a 24-word mnemonic.
// It uses CRC32 of the first prefixLen characters (or runes for UTF-8) of each word.
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/mnemonics/electrum-words.cpp#L186-L210
func calculateChecksumWithPrefix(words []string, prefixLen int) string {
	var prefixes strings.Builder
	for _, word := range words {
		// For UTF-8 languages (Chinese, Japanese, Russian), we need to count runes, not bytes
		runeCount := utf8.RuneCountInString(word)
		if runeCount >= prefixLen {
			count := 0
			for _, r := range word {
				prefixes.WriteRune(r)
				count++
				if count >= prefixLen {
					break
				}
			}
		} else {
			prefixes.WriteString(word)
		}
	}

	checksum := crc32.ChecksumIEEE([]byte(prefixes.String()))
	idx := checksum % 24

	return words[idx]
}

// Mnemonic returns the 25-word mnemonic seed for this key in the specified language.
func (k *Key) Mnemonic(language string) (string, error) {
	return KeyToMnemonic(k.PrivateSpendKey, language)
}

// NewKeyFromMnemonic creates a Key from a 25-word mnemonic seed.
// It automatically detects the language from the mnemonic words.
func NewKeyFromMnemonic(mnemonic string, network byte) (*Key, error) {
	spendKey, _, err := MnemonicToKey(mnemonic)
	if err != nil {
		return nil, err
	}

	k := &Key{}
	k.PrivateSpendKey = spendKey
	k.PrivateViewKey = deriveViewKey(spendKey[:])
	k.PublicSpendKey = publicKeyFromPrivateKey(k.PrivateSpendKey[:])
	k.PublicViewKey = publicKeyFromPrivateKey(k.PrivateViewKey[:])
	k.Address = encodeAddress(network, k.PublicSpendKey[:], k.PublicViewKey[:])

	return k, nil
}

// AvailableLanguages returns a list of available mnemonic language names.
func AvailableLanguages() []string {
	languages := make([]string, 0, len(Languages))
	for name := range Languages {
		languages = append(languages, name)
	}
	return languages
}
