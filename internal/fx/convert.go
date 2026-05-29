// Package fx faz conversão de moeda e provê taxas de câmbio.
package fx

import "math/big"

const rateScale = 100_000_000 // 1e8

// Convert calcula amountBase * rateE8 / 1e8 com floor, usando big.Int (sem overflow).
// rateE8 = unidades da moeda quote por 1 unidade base, escalado por 1e8.
func Convert(amountBase, rateE8 int64) int64 {
	prod := new(big.Int).Mul(big.NewInt(amountBase), big.NewInt(rateE8))
	prod.Quo(prod, big.NewInt(rateScale)) // Quo trunca em direção a zero (floor para positivos)
	return prod.Int64()
}
