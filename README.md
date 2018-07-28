# Shamir Secret Sharing Algorithm

This is my personal fork of:

An implementation of Shamir's Secret Sharing Algorithm in Go  

    Copyright (C) 2015 Alexander Scheel, Joel May, Matthew Burket  
    See Contributors.md for a complete list of contributors.  
    Licensed under the MIT License.  

## Usage
Secret sharing is done on byte-strings (smaller than 256-bit finite field of prime order) encoded directly. Shares are pairs of (x,y) points which can reconstruct the polynomial with lagrange polynomial interpolation. The idea of automatically splitting and encoding a larger than 256-bit secret has been completely extracted from the parent package to make it simpler. The same functionality could easily be reapplied on top of this library however.

```
shares := shamir.Shares(threshold int, total int, raw []byte) // creates a set of shares

shamir.Reconstruct(shares []*Share) // combines shares into secret

// This is what the Share struct looks like
type Share struct {
    X *big.Int
    Y *big.Int
}
```

Functions for the easy utilization of additive homomorphic properties has been put into the package, as well as the basic underlying primitives necessary for multiplication gates. Some helpers for interpolating and reconstructing polynomials with inverted vandermonde matrices is also added here -- though in basic reconstruction of P(0) the Horner Method is used (because it's faster).