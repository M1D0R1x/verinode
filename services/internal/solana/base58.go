package solana

import (
	"fmt"
	"math/big"
)

// base58.go — minimal Base58 (Bitcoin alphabet) codec, dependency-free.
// Solana addresses, blockhashes and signatures are all Base58-encoded.

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var base58Index = func() map[byte]int {
	m := make(map[byte]int, len(base58Alphabet))
	for i := 0; i < len(base58Alphabet); i++ {
		m[base58Alphabet[i]] = i
	}
	return m
}()

func base58Encode(input []byte) string {
	if len(input) == 0 {
		return ""
	}
	// Count leading zero bytes -> leading '1's.
	zeros := 0
	for zeros < len(input) && input[zeros] == 0 {
		zeros++
	}

	num := new(big.Int).SetBytes(input)
	radix := big.NewInt(58)
	mod := new(big.Int)
	var out []byte
	for num.Sign() > 0 {
		num.DivMod(num, radix, mod)
		out = append(out, base58Alphabet[mod.Int64()])
	}
	for i := 0; i < zeros; i++ {
		out = append(out, base58Alphabet[0])
	}
	// Reverse.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

func base58Decode(input string) ([]byte, error) {
	if input == "" {
		return []byte{}, nil
	}
	num := big.NewInt(0)
	radix := big.NewInt(58)
	for i := 0; i < len(input); i++ {
		idx, ok := base58Index[input[i]]
		if !ok {
			return nil, fmt.Errorf("invalid base58 character %q", input[i])
		}
		num.Mul(num, radix)
		num.Add(num, big.NewInt(int64(idx)))
	}
	decoded := num.Bytes()

	// Restore leading zero bytes for each leading '1'.
	zeros := 0
	for zeros < len(input) && input[zeros] == base58Alphabet[0] {
		zeros++
	}
	if zeros > 0 {
		decoded = append(make([]byte, zeros), decoded...)
	}
	return decoded, nil
}
