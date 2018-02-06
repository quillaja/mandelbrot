package big

import (
	"fmt"
	"math"
	"math/big"
)

// PrecisionRequired determines the rough minimum number bits of precision
// to hold a n-digit number.
func PrecisionRequired(digits int) uint {
	return uint(math.Ceil(float64(digits) * math.Log2(10)))
}

// Complex represents a complex number with arbitrary precision. It uses 2
// big.Float types, one `R` for the real part and one `I` for the imaginary
// part.
//
// Instantiation can be performed in the usual ways.
//     c1 := NewComplex(1.0,-10.25, 128) // 128 bits of precision
//     c2 := new(Complex) // default big.Float precision of 0
//     c3 := Complex{} // default big.Float precision of 0
//     var c4 Complex // default big.Float precision of 0
//
// Mathmatical operations generally follow the conventions of the standard
// library's `big` package. Notable exceptions are the Abs() and AbsSq()
// functions.
//
// The type's string representation mostly matches the built-in complex64 and
// complex128's format, except that there is ALWAYS a '+' between the real and
// imaginary components. For example, the built-in shows
//     (1.5-2.75i)
// for number with a negative imaginary part, whereas Complex will show
//     (1.5+-2.75i)
type Complex struct {
	R big.Float
	I big.Float
}

// NewComplex returns a pointer to a complex number with the given real and
// imaginary parts, both with `prec` precision.
func NewComplex(real, imag float64, prec uint) *Complex {
	return &Complex{
		*new(big.Float).SetPrec(prec).SetFloat64(real),
		*new(big.Float).SetPrec(prec).SetFloat64(imag)}
}

// Copy will copy a's value, precision, etc into z unless z==a.
func (z *Complex) Copy(a *Complex) *Complex {
	if z != a {
		z.R.Copy(&a.R)
		z.I.Copy(&a.I)
	}
	return z
}

// Add adds a and b and stores the result in z.
func (z *Complex) Add(a, b *Complex) *Complex {
	z.R.Add(&a.R, &b.R)
	z.I.Add(&a.I, &b.I)
	return z
}

// Sub subtracts b from a and stores the result in z.
func (z *Complex) Sub(a, b *Complex) *Complex {
	z.R.Sub(&a.R, &b.R)
	z.I.Sub(&a.I, &b.I)
	return z
}

// Mul multiplies a and b and stores the result in z.
func (z *Complex) Mul(a, b *Complex) *Complex {
	left, right := new(big.Float), new(big.Float)
	real := new(big.Float).Sub(left.Mul(&a.R, &b.R), right.Mul(&a.I, &b.I))
	z.I.Add(left.Mul(&a.R, &b.I), right.Mul(&a.I, &b.R))
	z.R.Copy(real)
	return z
}

// Div divides a by b and stores the result in z.
func (z *Complex) Div(a, b *Complex) *Complex {
	left, right := new(big.Float), new(big.Float)
	bottom := b.AbsSq()

	real := new(big.Float).Add(left.Mul(&a.R, &b.R), right.Mul(&a.I, &b.I)).Quo(&z.R, bottom)
	z.I.Sub(left.Mul(&a.I, &b.R), right.Mul(&a.R, &b.I)).Quo(&z.I, bottom)
	z.R.Copy(real)
	return z
}

// Pow2 squares a and stores the result in z. This is just Mul(a,a).
func (z *Complex) Pow2(a *Complex) *Complex {
	return z.Mul(a, a)
}

// AbsSq returns the |z|^2, which is useful for performance critical work
// since it eliminates a call to Sqrt().
func (z *Complex) AbsSq() *big.Float {
	left, right := new(big.Float), new(big.Float)
	left.Add(left.Mul(&z.R, &z.R), right.Mul(&z.I, &z.I))
	return left
}

// Abs returns the absolute value of z, |z|. The conversion to float64 may
// introduce error, but is necessary to use math.Sqrt(float64) since big.Float
// presently lacks a Sqrt() method (until Go v1.10?).
func (z *Complex) Abs() float64 {
	a, _ := z.AbsSq().Float64()
	return math.Sqrt(a)
}

// Neg negates a and stores the result in z.
func (z *Complex) Neg(a *Complex) *Complex {
	z.R.Neg(&a.R)
	z.I.Neg(&a.I)
	return z
}

// SetFloat64 sets the real and imaginary parts of the number to the parameters
// using the same rules as big.Float.SetFloat64().
func (z *Complex) SetFloat64(real, imag float64) *Complex {
	z.R.SetFloat64(real)
	z.I.SetFloat64(imag)
	return z
}

// SetComplex128 sets the real and imaginary compontents equal to those of the
// parameter, using the rules of SetFloat64().
func (z *Complex) SetComplex128(c complex128) *Complex {
	z.SetFloat64(real(c), imag(c))
	return z
}

// SetString sets the real and imaginary components to the values of the string
// parameters using big.Float.Parse().
func (z *Complex) SetString(real, imag string) *Complex {
	z.R.Parse(real, 0)
	z.I.Parse(imag, 0)
	return z
}

// Prec returns the lower of R.Prec() and I.Prec().
//
// Although probably not usual, R and I's precision may be set independently.
func (z *Complex) Prec() uint {
	if z.R.Prec() < z.I.Prec() {
		return z.R.Prec()
	}
	return z.I.Prec()
}

// SetPrec sets both R and I's precisions to `prec`.
func (z *Complex) SetPrec(prec uint) *Complex {
	z.R.SetPrec(prec)
	z.I.SetPrec(prec)
	return z
}

// Complex128 returns a `complex128`, with whatever precision losses happen
// from big.Float.Float64().
func (z *Complex) Complex128() complex128 {
	r, _ := z.R.Float64()
	i, _ := z.I.Float64()
	return complex(r, i)
}

// String returns a string representation of the complex number.
//
// See type documentation for other details.
func (z Complex) String() string {
	return fmt.Sprintf("(%g+%gi)", &z.R, &z.I)
}
