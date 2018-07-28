package shamir

import (
	"errors"

	"github.com/superarius/shamir/modular"
)

type Share struct {
	X *modular.Int
	Y *modular.Int
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

func (share *Share) ScalarMul(x *Share, n *modular.Int) *Share {
	share.X = x.X
	share.Y = new(modular.Int).Mul(x.Y, n)
	return share
}

func (share *Share) ScalarAdd(x *Share, n *modular.Int) *Share {
	num := new(modular.Int).Add(x.Y, n)
	share.X = x.X
	share.Y = num
	return share
}

type Triple struct {
	A *Share
	B *Share
	C *Share
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
	c := new(modular.Int).Mul(a, b)
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

func PrepareMul(x, y *Share, triple *Triple) (*Share, *Share) {
	neg1 := modular.NewInt(-1)
	negA := new(Share).ScalarMul(triple.A, neg1)
	negB := new(Share).ScalarMul(triple.B, neg1)
	xas := new(Share).Add(x, negA)
	ybs := new(Share).Add(y, negB)
	return xas, ybs
}

func FinishMul(x_a []*Share, y_b []*Share, x, y *Share, triple *Triple) (*Share, error) {
	epsilon, err := Reconstruct(x_a, len(x_a))
	if err != nil {
		return nil, err
	}
	rho, err := Reconstruct(y_b, len(y_b))
	if err != nil {
		return nil, err
	}
	ner := new(modular.Int).Mul(epsilon, rho)
	ner.Mul(ner, modular.NewInt(-1))
	term1 := new(Share).ScalarMul(y, epsilon)
	term2 := new(Share).ScalarMul(x, rho)
	t3 := new(Share).Add(term1, term2, triple.C)
	out := new(Share).ScalarAdd(t3, ner)
	return out, nil
}