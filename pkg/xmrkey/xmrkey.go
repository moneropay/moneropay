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
	"crypto/rand"

	"filippo.io/edwards25519"
	"golang.org/x/crypto/sha3"
)

// Network bytes for Monero standard address types (CRYPTONOTE_PUBLIC_ADDRESS_BASE58_PREFIX).
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/cryptonote_config.h#L227
const (
	NetworkMainnet  byte = 18 // addresses start with '4' (L227)
	NetworkTestnet  byte = 53 // addresses start with '9' (L270)
	NetworkStagenet byte = 24 // addresses start with '5' (L285)
)

// Key holds all the Monero key material
type Key struct {
	PrivateSpendKey [32]byte
	PrivateViewKey  [32]byte
	PublicSpendKey  [32]byte
	PublicViewKey   [32]byte
	Address         string
}

// NewKey generates a new random Monero key pair for the specified network
func NewKey(network byte) (*Key, error) {
	var seed [32]byte
	if _, err := rand.Read(seed[:]); err != nil {
		return nil, err
	}

	k := &Key{}

	// Private spend key is the seed after scalar reduction
	copy(k.PrivateSpendKey[:], scReduce32(seed[:]))

	// Private view key is derived by hashing the spend key
	k.PrivateViewKey = deriveViewKey(k.PrivateSpendKey[:])

	// Generate public keys from private keys
	k.PublicSpendKey = publicKeyFromPrivateKey(k.PrivateSpendKey[:])
	k.PublicViewKey = publicKeyFromPrivateKey(k.PrivateViewKey[:])

	// Generate address for the specified network
	k.Address = encodeAddress(network, k.PublicSpendKey[:], k.PublicViewKey[:])

	return k, nil
}

// keccak256 computes the Keccak-256 hash (legacy, not SHA3-256).
// Monero uses Keccak-256, not the NIST-standardized SHA3-256.
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/crypto/keccak.c#L119-L167
func keccak256(data ...[]byte) [32]byte {
	h := sha3.NewLegacyKeccak256()
	for _, d := range data {
		h.Write(d)
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

// scReduce32 reduces a 32-byte value modulo the ed25519 scalar order l.
// This is equivalent to Monero's sc_reduce32.
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/crypto/crypto-ops.c#L2433-L2546
func scReduce32(s []byte) []byte {
	var wide [64]byte
	copy(wide[:32], s)
	scalar, _ := edwards25519.NewScalar().SetUniformBytes(wide[:])
	return scalar.Bytes()
}

// deriveViewKey derives the private view key from the private spend key.
// view_key = sc_reduce32(keccak256(spend_key))
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/crypto/crypto.cpp#L148-L173
func deriveViewKey(spendKey []byte) [32]byte {
	hash := keccak256(spendKey)
	var result [32]byte
	copy(result[:], scReduce32(hash[:]))
	return result
}

// publicKeyFromPrivateKey computes the public key from a private key.
// public_key = private_key * G (scalar multiplication with base point)
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/crypto/crypto-ops.c#L1639-L1679
func publicKeyFromPrivateKey(privateKey []byte) [32]byte {
	// Private key is already a reduced scalar, so use SetCanonicalBytes
	scalar, _ := edwards25519.NewScalar().SetCanonicalBytes(privateKey)
	point := edwards25519.NewGeneratorPoint().ScalarBaseMult(scalar)

	var result [32]byte
	copy(result[:], point.Bytes())
	return result
}

// encodeAddress creates a Monero address from the public keys.
// Address format: network_byte || public_spend_key || public_view_key || checksum
// Checksum is first 4 bytes of keccak256(data).
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/common/base58.cpp#L227-L235
func encodeAddress(network byte, publicSpendKey, publicViewKey []byte) string {
	data := make([]byte, 0, 69)
	data = append(data, network)
	data = append(data, publicSpendKey...)
	data = append(data, publicViewKey...)

	hash := keccak256(data)
	data = append(data, hash[:4]...)

	return base58Encode(data)
}

// Monero base58 alphabet (same as Bitcoin).
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/common/base58.cpp#L47
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// encodedBlockSizes maps input byte count to output character count.
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/common/base58.cpp#L49
var encodedBlockSizes = [9]int{0, 2, 3, 5, 6, 7, 9, 10, 11}

// base58Encode encodes data using Monero's base58 variant.
// Unlike standard base58, Monero encodes in 8-byte blocks to 11-character chunks.
// This ensures fixed-length encoding and easier checksum validation.
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/common/base58.cpp#L172-L193
func base58Encode(data []byte) string {
	var result []byte

	fullBlocks := len(data) / 8
	remainder := len(data) % 8

	for i := range fullBlocks {
		block := data[i*8 : (i+1)*8]
		encoded := encodeBlock(block, 11)
		result = append(result, encoded...)
	}

	if remainder > 0 {
		block := data[fullBlocks*8:]
		encoded := encodeBlock(block, encodedBlockSizes[remainder])
		result = append(result, encoded...)
	}

	return string(result)
}

// encodeBlock converts a block of bytes to base58 characters.
// See: https://github.com/monero-project/monero/blob/v0.18.4.5/src/common/base58.cpp#L123-L136
func encodeBlock(block []byte, encodedLen int) []byte {
	var num uint64
	for _, b := range block {
		num = num*256 + uint64(b)
	}

	result := make([]byte, encodedLen)
	for i := encodedLen - 1; i >= 0; i-- {
		result[i] = base58Alphabet[num%58]
		num /= 58
	}

	return result
}
