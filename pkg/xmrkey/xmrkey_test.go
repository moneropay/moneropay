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
	"encoding/hex"
	"strings"
	"testing"
)

func TestNewKey(t *testing.T) {
	key, err := NewKey(NetworkMainnet)
	if err != nil {
		t.Fatalf("NewKey() failed: %v", err)
	}

	// Check address format (should start with 4 for mainnet)
	if !strings.HasPrefix(key.Address, "4") {
		t.Errorf("Address should start with '4' for mainnet, got: %s", key.Address)
	}

	// Check address length (standard Monero address is 95 characters)
	if len(key.Address) != 95 {
		t.Errorf("Address should be 95 characters, got: %d", len(key.Address))
	}

	// Check private keys are 32 bytes
	if len(key.PrivateSpendKey) != 32 {
		t.Errorf("Private spend key should be 32 bytes, got: %d", len(key.PrivateSpendKey))
	}
	if len(key.PrivateViewKey) != 32 {
		t.Errorf("Private view key should be 32 bytes, got: %d", len(key.PrivateViewKey))
	}
}

func TestUniqueKeys(t *testing.T) {
	key1, _ := NewKey(NetworkMainnet)
	key2, _ := NewKey(NetworkMainnet)

	if key1.Address == key2.Address {
		t.Error("Two generated keys should have different addresses")
	}
}

// Test vectors from Monero source code
// https://github.com/monero-project/monero/blob/v0.18.4.5/tests/unit_tests/base58.cpp#L174-L206
func TestBase58EncodeBlock(t *testing.T) {
	tests := []struct {
		input    []byte
		expected string
	}{
		{[]byte{0x00}, "11"},
		{[]byte{0x39}, "1z"},
		{[]byte{0xFF}, "5Q"},
		{[]byte{0x00, 0x00}, "111"},
		{[]byte{0x00, 0x39}, "11z"},
		{[]byte{0x01, 0x00}, "15R"},
		{[]byte{0xFF, 0xFF}, "LUv"},
		{[]byte{0x00, 0x00, 0x00}, "11111"},
		{[]byte{0x00, 0x00, 0x39}, "1111z"},
		{[]byte{0x01, 0x00, 0x00}, "11LUw"},
		{[]byte{0xFF, 0xFF, 0xFF}, "2UzHL"},
		{[]byte{0x00, 0x00, 0x00, 0x39}, "11111z"},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF}, "7YXq9G"},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x39}, "111111z"},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, "VtB5VXc"},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x39}, "11111111z"},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, "3CUsUpv9t"},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x39}, "111111111z"},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, "Ahg1opVcGW"},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x39}, "1111111111z"},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, "jpXCZedGfVQ"},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, "11111111111"},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "11111111112"},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08}, "11111111119"},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09}, "1111111111A"},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x3A}, "11111111121"},
		{[]byte{0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, "1Ahg1opVcGW"},
		{[]byte{0x06, 0x15, 0x60, 0x13, 0x76, 0x28, 0x79, 0xF7}, "22222222222"},
		{[]byte{0x05, 0xE0, 0x22, 0xBA, 0x37, 0x4B, 0x2A, 0x00}, "1z111111111"},
	}

	for _, tt := range tests {
		result := string(encodeBlock(tt.input, encodedBlockSizes[len(tt.input)]))
		if result != tt.expected {
			t.Errorf("encodeBlock(%x) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

// Test vectors from Monero source code
// https://github.com/monero-project/monero/blob/v0.18.4.5/tests/unit_tests/base58.cpp#L286-L302
func TestBase58Encode(t *testing.T) {
	tests := []struct {
		input    []byte
		expected string
	}{
		{[]byte{0x00}, "11"},
		{make([]byte, 2), "111"},
		{make([]byte, 3), "11111"},
		{make([]byte, 4), "111111"},
		{make([]byte, 5), "1111111"},
		{make([]byte, 6), "111111111"},
		{make([]byte, 7), "1111111111"},
		{make([]byte, 8), "11111111111"},
		{make([]byte, 9), "1111111111111"},
		{make([]byte, 10), "11111111111111"},
		{make([]byte, 11), "1111111111111111"},
		{make([]byte, 12), "11111111111111111"},
		{make([]byte, 13), "111111111111111111"},
		{make([]byte, 14), "11111111111111111111"},
		{make([]byte, 15), "111111111111111111111"},
		{make([]byte, 16), "1111111111111111111111"},
		{[]byte{0x06, 0x15, 0x60, 0x13, 0x76, 0x28, 0x79, 0xF7,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, "22222222222VtB5VXc"},
	}

	for _, tt := range tests {
		result := base58Encode(tt.input)
		if result != tt.expected {
			t.Errorf("base58Encode(%x) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestEncodeAddressVector(t *testing.T) {
	// Test vector from Monero unit tests
	// See: https://github.com/monero-project/monero/blob/v0.18.4.5/tests/unit_tests/base58.cpp#L464-L470
	publicSpendKey, _ := hex.DecodeString("f724bc5c6cfbb9d97602c300423a2f28641874513a035778a0c1778d833201e9")
	publicViewKey, _ := hex.DecodeString("220939689edf1abd5bc1d031f73ecd6c993add66d6808870456afeb8e7eeb68d")

	// Monero mainnet tag is 18
	expected := "4AzKEX4gXdJdNeM6dfiBFL7kqund3HYGvMBF3ttsNd9SfzgYB6L7ep1Yg1osYJzLdaKAYSLVh6e6jKnAuzj3bw1oGy9kXCb"

	result := encodeAddress(NetworkMainnet, publicSpendKey, publicViewKey)
	if result != expected {
		t.Errorf("encodeAddress() = %s, want %s", result, expected)
	}
}

func TestNewKeyNetworks(t *testing.T) {
	tests := []struct {
		network      byte
		expectedChar string
		name         string
	}{
		{NetworkMainnet, "4", "mainnet"},
		{NetworkTestnet, "9", "testnet"},
		{NetworkStagenet, "5", "stagenet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := NewKey(tt.network)
			if err != nil {
				t.Fatalf("NewKey(%s) failed: %v", tt.name, err)
			}
			if !strings.HasPrefix(key.Address, tt.expectedChar) {
				t.Errorf("Address should start with '%s' for %s, got: %s",
					tt.expectedChar, tt.name, key.Address)
			}
		})
	}
}

func TestEncodeAddressTestnet(t *testing.T) {
	// Test vector from Monero serialization tests
	// See: https://github.com/monero-project/monero/blob/v0.18.4.5/tests/unit_tests/serialization.cpp#L770-L772
	publicSpendKey, _ := hex.DecodeString("9bc53a6ff7b0831c9470f71b6b972dbe5ad1e8606f72682868b1dda64e119fb3")
	publicViewKey, _ := hex.DecodeString("49fece1ef97dc0c0f7a5e2106e75e96edd910f7e86b56e1e308cd0cf734df191")

	expected := "9y52S6PH8tz5nG6m7NhaD2YqgRRKfYdJ77m226yoPJKQWzL3TPWh6WjZH2Ty3g2m4tKYXyxpgcssB63stc5JQudJHUBA2RA"

	result := encodeAddress(NetworkTestnet, publicSpendKey, publicViewKey)
	if result != expected {
		t.Errorf("encodeAddress() = %s, want %s", result, expected)
	}
}

// Test vectors generated using monero-wallet-cli v0.18.4.5
// These wallets were created specifically for testing with "test123" password.
// Each test verifies the full key derivation chain: spend key -> view key -> public keys -> address

func TestWalletCLIVectors(t *testing.T) {
	tests := []struct {
		name                   string
		network                byte
		privateSpendKey        string
		expectedPrivateViewKey string
		expectedPublicSpendKey string
		expectedPublicViewKey  string
		expectedAddress        string
	}{
		{
			name:                   "mainnet",
			network:                NetworkMainnet,
			privateSpendKey:        "83aaa12a23059f50c63e1a4b11589675aef480122951a66f7eeeaaffba16060e",
			expectedPrivateViewKey: "6f08345e99e276ef85c0bf0218f7c600593565b09b600f9a3de8e20760adf20c",
			expectedPublicSpendKey: "2b016045d9f106f602aeff9f0bc832618e43a707b2bb06056c0be863ae3ef912",
			expectedPublicViewKey:  "b40df14547f12b301ec4f6e63d41ed6c6b32792c3dd94d032a8cf9be7504c9c3",
			expectedAddress:        "43FdbhX4svdi9cNXgKoei5HKQwGrBXdrm1ubqtxVeJJL48Srgt9Qefc93pvAeNk1ELK8oG46NSzHJ1XiYRgAjowNP7Yc6p4",
		},
		{
			name:                   "testnet",
			network:                NetworkTestnet,
			privateSpendKey:        "59cfae3d4d6a99038b7844c5d61b5d6350077423f013512c19232f37669dcc05",
			expectedPrivateViewKey: "a90811463f72812b957717cc8223f449df338f71009fbcb8435451b1776a1105",
			expectedPublicSpendKey: "1e22573678357e97d15e7be488f1909283b96f4e84dc8f6a2a807cf23b1bcca5",
			expectedPublicViewKey:  "c8670bfa3f37a75a6ba046aaf7fb6f8800cbe117a13368ba8c23038136826bec",
			expectedAddress:        "9tJtFjPZNDsSPpkahxcrmDRWNjKetycziJkwie9nvgoDUjJTiVLSYo8G8C6UvFrmwUPkQERHZSAb9YCk8JGVGpRgTgEF661",
		},
		{
			name:                   "stagenet",
			network:                NetworkStagenet,
			privateSpendKey:        "9710f183e594b61c93bdca6ea485de53a794cde8d955df1a7ab0d138b102c30e",
			expectedPrivateViewKey: "989ce17ac2a17fcb2a4c7a86c3140bdce91c9cf3fd0e532c24fd0a3dedb7e205",
			expectedPublicSpendKey: "c22c270797555b101665a394127bed9846e81194cf06d1dc085c6a652b8b3c4f",
			expectedPublicViewKey:  "0e789f04ef0f7ca70555e1eb21809b4978c5f4eb664f82f5619c7a51ef2c31a7",
			expectedAddress:        "59BvYSmW9AA3h4zQdc4F3SSUH6KJQRWDvdobMvjBaiYjEDwwtmpf7j1UwK4VeFH3L6DHmZFpF6F9si3WLwdgPCUgKxC58px",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateSpendKey, _ := hex.DecodeString(tt.privateSpendKey)

			// Test view key derivation
			viewKey := deriveViewKey(privateSpendKey)
			if hex.EncodeToString(viewKey[:]) != tt.expectedPrivateViewKey {
				t.Errorf("deriveViewKey() = %s, want %s", hex.EncodeToString(viewKey[:]), tt.expectedPrivateViewKey)
			}

			// Test public key derivation
			publicSpendKey := publicKeyFromPrivateKey(privateSpendKey)
			if hex.EncodeToString(publicSpendKey[:]) != tt.expectedPublicSpendKey {
				t.Errorf("publicKeyFromPrivateKey(spend) = %s, want %s", hex.EncodeToString(publicSpendKey[:]), tt.expectedPublicSpendKey)
			}

			publicViewKey := publicKeyFromPrivateKey(viewKey[:])
			if hex.EncodeToString(publicViewKey[:]) != tt.expectedPublicViewKey {
				t.Errorf("publicKeyFromPrivateKey(view) = %s, want %s", hex.EncodeToString(publicViewKey[:]), tt.expectedPublicViewKey)
			}

			// Test address encoding
			address := encodeAddress(tt.network, publicSpendKey[:], publicViewKey[:])
			if address != tt.expectedAddress {
				t.Errorf("encodeAddress() = %s, want %s", address, tt.expectedAddress)
			}
		})
	}
}
