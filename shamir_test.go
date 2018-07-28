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

func TestPolynomialAddition(t *testing.T) {
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

	answer := new(modular.Int).Add(sec1, sec2)

	result, err := Reconstruct(finalShares[:10], 10)
	require.NoError(err)
	require.Equal(0, answer.Cmp(result), "did not interpolate correctly")
}

func TestTripleCreation(t *testing.T) {
	require := require.New(t)

	// Test Triple Creation
	triples, err := NewTriples(100, 6, 19) // creates 100 triple sharings
	require.NoError(err)

	require.Equal(0, triples[0][0].A.X.Cmp(triples[0][3].B.X), "x values should be the same")
	require.Equal(0, triples[1][1].A.X.Cmp(triples[1][90].C.X), "x values should be the same")

	// Test Batched Triple Creation
	triple_groups, err := NewBatchedTriples(2, 2, 7) // creates 20k triple sharings (batched into 10k groups). When tested on a 6/19 network we can create 100k sharings in just under 30 seconds.
	require.NoError(err)
	require.Equal(0, triple_groups[0][0][0].A.X.Cmp(triple_groups[1][0][9000].B.X), "x values should be the same")	
}
