package shamir

import (

	"github.com/superarius/shamir/modular"
)


func EvaluatePolynomial(polynomial []*modular.Int, value *modular.Int) *modular.Int {
	last := len(polynomial) - 1
	result := modular.IntFromBig(polynomial[last].AsBig())

	for s := last - 1; s >= 0; s-- {
		result.Mul(result, value)
		result.Add(result, polynomial[s])
	}

	return result
}

func VandermondeRow(x *modular.Int, cols int) []*modular.Int {
	row := make([]*modular.Int, cols)
	for i := range row {
		val := new(modular.Int).Exp(x, modular.NewInt(int64(i)))
		row[i] = val
	}

	return row
}

func VandermondeSubset(xs []*modular.Int, cols int) *modular.Matrix {
	vdm := modular.NewMatrix(len(xs), cols, []*modular.Int{})
	for i, x := range xs {
		vdm.SetRow(i+1, VandermondeRow(x, cols))
	}

	return vdm
}

func InterpolatePolynomial(points []*Share) ([]*modular.Int, error) {
	xs := make([]*modular.Int, len(points))
	ys := make([]*modular.Int, len(points))
	for i, p := range points {
		xs[i] = p.X
		ys[i] = p.Y
	}
	m := VandermondeSubset(xs, len(xs))
	inv, err := m.Inverse()
	if err != nil {
		return nil, err
	}
	ym := modular.NewMatrix(len(ys), 1, ys)
	polym, err := new(modular.Matrix).Mul(inv, ym)
	poly := polym.GetRow(1)
	if err != nil {
		return nil, err
	}	

	return poly, nil
}

func InterpolateAtZero(points []*Share) *modular.Int {

	secret := modular.NewInt(0)
	for i, p := range points { 
		origin := p.X
		originy := p.Y
		numerator := modular.NewInt(1)  
		denominator := modular.NewInt(1) 
		for k := range points {
			if k != i {
				current := points[k].X
				negative := modular.NewInt(0)
				negative.Mul(current, modular.NewInt(-1))
				added := new(modular.Int).Sub(origin, current)
				numerator.Mul(numerator, negative)
				denominator.Mul(denominator, added)			
			}
		}

		working := modular.IntFromBig(originy.AsBig())
		working.Mul(working, numerator)
		working.Mul(working, modular.ModInverse(denominator))

		secret.Add(secret, working)
	}

	return secret
}