package big

import (
	"fmt"
	"math"
	"math/big"
)

type Complex struct {
	R big.Float
	I big.Float
}

func NewComplex(real, imag float64, prec uint) *Complex {
	return &Complex{
		*new(big.Float).SetPrec(prec).SetFloat64(real),
		*new(big.Float).SetPrec(prec).SetFloat64(imag)}
}

func (z *Complex) Copy(a *Complex) *Complex {
	if z != a {
		z.R.Copy(&a.R)
		z.I.Copy(&z.I)
	}
	return z
}

func (z *Complex) Add(a, b *Complex) *Complex {
	z.R.Add(&a.R, &b.R)
	z.I.Add(&a.I, &b.I)
	return z
}

func (z *Complex) Sub(a, b *Complex) *Complex {
	z.R.Sub(&a.R, &b.R)
	z.I.Sub(&a.I, &b.I)
	return z
}

func (z *Complex) Mul(a, b *Complex) *Complex {
	left, right := new(big.Float), new(big.Float)
	z.R.Sub(left.Mul(&a.R, &b.R), right.Mul(&a.I, &b.I))
	z.I.Add(left.Mul(&a.R, &b.I), right.Mul(&a.I, &b.R))
	return z
}

func (z *Complex) Div(a, b *Complex) *Complex {
	left, right := new(big.Float), new(big.Float)
	bottom := b.AbsSq()

	z.R.Add(left.Mul(&a.R, &b.R), right.Mul(&a.I, &b.I)).Quo(&z.R, bottom)
	z.I.Sub(left.Mul(&a.I, &b.R), right.Mul(&a.R, &b.I)).Quo(&z.I, bottom)
	return z
}

func (z *Complex) Pow2(a *Complex) *Complex {
	return z.Mul(a, a)
}

func (z *Complex) AbsSq() *big.Float {
	left, right := new(big.Float), new(big.Float)
	left.Add(left.Mul(&z.R, &z.R), right.Mul(&z.I, &z.I))
	return left
}

func (z *Complex) Abs() float64 {
	a, _ := z.AbsSq().Float64()
	return math.Sqrt(a)
}

func (z *Complex) Neg() *Complex {
	z.R.Neg(&z.R)
	z.I.Neg(&z.I)
	return z
}

func (z *Complex) SetFloat64(real, imag float64) *Complex {
	z.R.SetFloat64(real)
	z.I.SetFloat64(imag)
	return z
}

func (z *Complex) SetComplex128(c complex128) *Complex {
	z.SetFloat64(real(c), imag(c))
	return z
}

func (z *Complex) SetString(real, imag string) *Complex {
	z.R.Parse(real, 0)
	z.I.Parse(imag, 0)
	return z
}

func (z *Complex) Prec() uint {
	if z.R.Prec() < z.I.Prec() {
		return z.R.Prec()
	} else if z.R.Prec() > z.I.Prec() {
		return z.I.Prec()
	}
	return z.R.Prec()
}

func (z *Complex) SetPrec(prec uint) *Complex {
	z.R.SetPrec(prec)
	z.I.SetPrec(prec)
	return z
}

func (z *Complex) Complex128() complex128 {
	r, _ := z.R.Float64()
	i, _ := z.I.Float64()
	return complex(r, i)
}

func (z Complex) String() string {
	return fmt.Sprintf("(%g+%gi)", &z.R, &z.I)
}
