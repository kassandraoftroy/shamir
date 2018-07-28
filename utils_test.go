package shamir

import (
	"testing"

	"github.com/superarius/shamir/modular"
	"github.com/stretchr/testify/require"
)

func TestVandermonde(t *testing.T) {
	require := require.New(t)

	// Test Vandermonde Subset
	xs := []*modular.Int{modular.NewInt(1), modular.NewInt(2), modular.NewInt(4), modular.NewInt(5), modular.NewInt(10)}
	m := VandermondeSubset(xs, 5)
	check := m.Represent2D()
	require.Equal(0, check[0][0].Cmp(check[1][0]), "vandermonde failed")
	require.Equal(0, check[4][4].Cmp(modular.NewInt(10000)), "vandermonde failed")

	// Test Interpolate Polynomial
	poly := make([]*modular.Int, 10)
	for i := range poly {
		v, err := modular.RandInt()
		require.NoError(err)
		poly[i] = v
	}

	points := make([]*Share, 11)
	for i := range points {
		x := modular.NewInt(int64(i+1))
		points[i] = &Share{
			X: x,
			Y: EvaluatePolynomial(poly, x),
		}
	}

	res, err := InterpolatePolynomial(points[:10])
	require.NoError(err)
	require.Equal(0, poly[0].Cmp(res[0]), "vandermond interpolation failed")
	require.Equal(0, poly[4].Cmp(res[4]), "vandermond interpolation failed")
}