# Shamir Secret Sharing Algorithm

This is my personal fork of:

An implementation of Shamir's Secret Sharing Algorithm in Go  

    Copyright (C) 2015 Alexander Scheel, Joel May, Matthew Burket  
    See Contributors.md for a complete list of contributors.  
    Licensed under the MIT License.  

The repo has been changed in a number of ways:
- Added a modular math package to wrap big.Int and directly do modular reductions
- Simplified secret sharing to only handle messages under 256 bits leaving chunking of longer messages out (or up to user)
- Added funtionality to interpolate via vandermonde matrices
- Added some functions to help with circuit evaluation over (share addition, and beaver triples for multiplicative gates)


## Usage
Secret sharing is done on byte-strings (smaller than 256-bit finite field of prime order) encoded directly. Shares are pairs of (x,y) points which can reconstruct the polynomial with lagrange polynomial interpolation. Basic splitting and reconstruction:

```
shares := shamir.Shares(threshold int, total int, raw []byte) // creates a set of shares

shamir.Reconstruct(shares []*Share) // combines shares into secret

// This is what the Share struct looks like
type Share struct {
    X *modular.Int
    Y *modular.Int
}
```

Functions for the easy utilization of additive homomorphic properties has been put into the package, as well as the basic underlying primitives necessary for multiplication gates. Some helpers for interpolating and reconstructing polynomials with inverted vandermonde matrices is also added here -- though in basic reconstruction of P(0) the Horner Method is used (because it's faster).