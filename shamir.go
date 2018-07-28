package shamir

import (
	"errors"

	"github.com/superarius/shamir/modular"
)

type Share struct {
	X *modular.Int
	Y *modular.Int
}

type Triple struct {
	A *Share
	B *Share
	C *Share
}


func Shares(threshold, total int, raw []byte) ([]*Share, error) {

	if threshold > total {
		return nil, errors.New("cannot require more shares then existing")
	}

	// Convert the secret to modular Int
	secret := modular.IntFromBytes(raw)

	if !secret.IsModP() {
		return nil, errors.New("secret is too large to encrypt")
	}

	// Create the polynomial of degree (threshold - 1)
	polynomial := make([]*modular.Int, threshold)
	polynomial[0] = secret
	for j := range polynomial[1:] {
		r, err := modular.RandInt()
		if err != nil {
			return nil, err
		}
		polynomial[j+1] = r
	}

	// Create the (x, y) points of each share.
	result := make([]*Share, total)

	for i := range result {
		// x-coordinate (taken as Natural Numbers 1,2,3...)
		x := modular.NewInt(int64(i+1))

		// evaluate the random polynomial at x
		result[i] = &Share{
			X: x,
			Y: EvaluatePolynomial(polynomial, x),
		}
	}

	return result, nil
}


func Reconstruct(shares []*Share, threshold int) (*modular.Int, error) {

	if len(shares) < threshold {
		return nil, errors.New("not enough shares for interoplation")
	}

	if len(shares) > threshold {
		shares = shares[:threshold]
	}

	// Use Lagrange Polynomial Interpolation to reconstruct the secret.
	secret := InterpolateAtZero(shares)

	return secret, nil
}

func (share *Share) Add(shares ...*Share) *Share {
	share.X = shares[0].X
	share.Y = modular.NewInt(0)
	for _, s := range shares {
		share.Y.Add(share.Y, s.Y)
	}
	return share
}

func TripleShares(t, n int) ([]*Triple, error) {
	a, err := modular.RandInt()
	if err != nil {
		return nil, err
	}
	b, err := modular.RandInt()
	if err != nil {
		return nil, err
	}
	c := new(modular.Int).Add(a, b)
	ashares, err := Shares(t, n, a.Bytes())
	if err != nil {
		return nil, err
	}
	bshares, err := Shares(t, n, b.Bytes())
	if err != nil {
		return nil, err
	}
	cshares, err := Shares(t, n, c.Bytes())
	if err != nil {
		return nil, err
	}
	triples := make([]*Triple, n)
	for i := range triples {
		triples[i] = &Triple {
			A: ashares[i],
			B: bshares[i],
			C: cshares[i],
		}
	}

	return triples, nil
}

func NewTriples(triples int, t, n int) ([][]*Triple, error) {
	i := 0 
	all := make([][]*Triple, triples)
	for i < triples {
		ts, err := TripleShares(t, n)
		if err != nil {
			return nil, err
		}
		all[i] = ts
		i++
	}
	out := make([][]*Triple, n)
	for i := range out {
		out[i] = make([]*Triple, triples)
		for j, a := range all {
			out[i][j] = a[i]
		}
	}
	return out, nil
}

func NewBatchedTriples(ten_ks int, t, n int) ([][][]*Triple, error) {
	i := 0
	batches := make([][][]*Triple, ten_ks)
	for i < ten_ks {
		ts, err := NewTriples(10000, t, n)
		if err != nil {
			return nil, err
		}
		batches[i] = ts
		i++
	}
	return batches, nil
}