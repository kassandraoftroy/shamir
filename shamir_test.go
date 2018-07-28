package shamir

import (
	"testing"

	"github.com/superarius/shamir/modular"
	"github.com/stretchr/testify/require"
)

func TestShareReconstruct(t *testing.T) {
	require := require.New(t)

	threshold := []int{2, 4, 6, 10, 20, 50, 100}
	shares := []int{7, 13, 19, 31, 61, 151, 301}

	// run tests with different threshold to share amounts
	for i := range threshold {
		secret, err := modular.RandInt()
		require.NoError(err)

		created, err := Shares(threshold[i], shares[i], secret.Bytes())
		require.NoError(err)

		combined, err := Reconstruct(created[:threshold[i]], threshold[i])
		require.NoError(err)

		require.Equal(combined, secret)
	}
}

func TestShareAdd(t *testing.T) {
	require := require.New(t)

	sec1, err := modular.RandInt()
	require.NoError(err)
	shares1, err := Shares(10, 31, sec1.Bytes())
	require.NoError(err)

	sec2, err := modular.RandInt()
	require.NoError(err)
	shares2, err := Shares(10, 31, sec2.Bytes())
	require.NoError(err)

	finalShares := make([]*Share, len(shares1))
	for i := range finalShares {
		finalShares[i] = new(Share).Add(shares1[i], shares2[i])
	}

	check := new(modular.Int).Add(sec1, sec2)

	result, err := Reconstruct(finalShares[:10], 10)
	require.NoError(err)
	require.Equal(0, check.Cmp(result), "did not interpolate correctly")
}

func TestShareScalarAdd(t *testing.T) {
	require := require.New(t)

	sec1, err := modular.RandInt()
	require.NoError(err)
	shares, err := Shares(10, 31, sec1.Bytes())
	require.NoError(err)

	finalShares	:= make([]*Share, 10)
	for i := range finalShares {
		finalShares[i] = new(Share).ScalarAdd(shares[i], modular.NewInt(1))
	}

	sec1.Add(sec1, modular.NewInt(1))

	result, err := Reconstruct(finalShares, 10)
	require.NoError(err)
	require.Equal(0, result.Cmp(sec1), "scalar addition failed")
}

func TestShareScalarMul(t *testing.T) {
	require := require.New(t)

	sec1, err := modular.RandInt()
	require.NoError(err)
	shares, err := Shares(10, 31, sec1.Bytes())
	require.NoError(err)

	finalShares	:= make([]*Share, 10)
	for i := range finalShares {
		finalShares[i] = new(Share).ScalarMul(shares[i], modular.NewInt(2))
	}

	sec1.Mul(sec1, modular.NewInt(2))

	result, err := Reconstruct(finalShares, 10)
	require.NoError(err)
	require.Equal(0, result.Cmp(sec1), "scalar addition failed")
}

func TestTripleCreation(t *testing.T) {
	require := require.New(t)

	// Test Single Triple Correctness
	threshold := 10
	n_nodes := 31
	trip, err := TripleShares(threshold, n_nodes)
	require.NoError(err)

	ashares := make([]*Share, n_nodes)
	bshares := make([]*Share, n_nodes)
	cshares := make([]*Share, n_nodes)
	for i, t := range trip {
		ashares[i] = t.A
		bshares[i] = t.B
		cshares[i] = t.C
	}

	a, err := Reconstruct(ashares, threshold)
	require.NoError(err)
	b, err := Reconstruct(bshares, threshold)
	require.NoError(err)
	c, err := Reconstruct(cshares, threshold)
	require.NoError(err)

	c2 := new(modular.Int).Mul(a, b)
	require.Equal(0, c2.Cmp(c), "triple is wrong")

	// Test Multiple Triple Creation
	triples, err := NewTriples(100, 6, 19) // creates 100 triple sharings
	require.NoError(err)

	require.Equal(0, triples[0][0].A.X.Cmp(triples[0][3].B.X), "x values should be the same")
	require.Equal(0, triples[1][1].A.X.Cmp(triples[1][90].C.X), "x values should be the same")

	// Test Batched Triple Creation
	triple_groups, err := NewBatchedTriples(2, 2, 7) // creates 20k triple sharings (batched into 10k groups). When tested on a 6/19 network we can create 100k sharings in just under 30 seconds.
	require.NoError(err)
	require.Equal(0, triple_groups[0][0][0].A.X.Cmp(triple_groups[1][0][9000].B.X), "x values should be the same")
	
}

func TestMultiplication(t *testing.T) {
	require := require.New(t)
	threshold := 10
	n_nodes := 31

	x, err := modular.RandInt()
	require.NoError(err)
	xshares, err := Shares(threshold, n_nodes, x.Bytes())
	require.NoError(err)

	y, err := modular.RandInt()
	require.NoError(err)
	yshares, err := Shares(threshold, n_nodes, y.Bytes())
	require.NoError(err)

	check := new(modular.Int).Mul(x, y)

	trip, err := TripleShares(threshold, n_nodes)
	require.NoError(err)

	xashares := make([]*Share, n_nodes)
	ybshares := make([]*Share, n_nodes)
	for i, t := range trip {
		xas, ybs := PrepareMul(xshares[i], yshares[i], t)
		xashares[i] = xas
		ybshares[i] = ybs
	}

	finalShares := make([]*Share, n_nodes)
	for i := range finalShares {
		s, err := FinishMul(xashares[:threshold], ybshares[:threshold], xshares[i], yshares[i], trip[i])
		require.NoError(err)
		finalShares[i] = s
	}

	result, err := Reconstruct(finalShares, threshold)
	require.NoError(err)
	require.Equal(0, result.Cmp(check), "multiplication failed")
}
