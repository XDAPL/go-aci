package aci

import (
	"fmt"
	"testing"
)

func TestSecurityStrengthFactor_codecov(t *testing.T) {
	var (
		factor SecurityStrengthFactor
		typ    string = BindSSF.String()
		err    error
	)

	for i := 0; i < 257; i++ {
		want := itoa(i) // what we expect (string representation)

		got := factor.Set(i) // set by iterated integer
		if want != got.String() {
			err = unexpectedStringResult(typ, want, got.String())
			t.Errorf("%s failed [int factor]: %v",
				t.Name(), err)
		}

		// reset using string representation of iterated integer
		if got = factor.Set(want); want != got.String() {
			err = unexpectedStringResult(typ, want, got.String())
			t.Errorf("%s failed [str factor]: %v",
				t.Name(), err)
		}

		// tod qualifies for all comparison operators
		// due to its numerical nature.
		cops := map[ComparisonOperator]func() BindRule{
			Eq: got.Eq,
			Ne: got.Ne,
			Lt: got.Lt,
			Le: got.Le,
			Gt: got.Gt,
			Ge: got.Ge,
		}

		// try every comparison operator supported in
		// this context ...
		for c := 1; c < len(cops)+1; c++ {
			cop := ComparisonOperator(c)
			wcop := sprintf("%s %s %q", got.Keyword(), cop, got)

			// create bindrule B using comparison
			// operator (cop).
			if B := cops[cop](); B.String() != wcop {
				err = unexpectedStringResult(typ, wcop, B.String())
			}

			if err != nil {
				t.Errorf("%s failed [factor rule]: %v", t.Name(), err)
			}
		}

	}

	// try to set our factor using special keywords
	// this package understands ...
	for word, value := range map[string]string{
		`mAx`:  `256`,
		`full`: `256`,
		`nOnE`: `0`,
		`OFF`:  `0`,
		`fart`: `0`,
	} {
		if got := factor.Set(word); got.String() != value {
			err = unexpectedStringResult(typ, value, got.String())
			t.Errorf("%s failed [factor word '%s']: %v", t.Name(), word, err)
		}
	}
}

func ExampleSecurityStrengthFactor_Set_byWordNoFactor() {
	var s SecurityStrengthFactor
	s.Set(`noNe`) // case is not significant
	fmt.Printf("%s", s)
	// Output: 0
}

func ExampleSecurityStrengthFactor_Set_byWordMaxFactor() {
	var s SecurityStrengthFactor
	s.Set(`FULL`) // case is not significant
	//s.Set(`max`) // alternative term
	fmt.Printf("%s", s)
	// Output: 256
}

func ExampleSecurityStrengthFactor_Set_byNumber() {
	var s SecurityStrengthFactor
	s.Set(128)
	fmt.Printf("%s\n", s)
	// Output: 128
}

func ExampleSecurityStrengthFactor_Eq() {
	var s SecurityStrengthFactor
	fmt.Printf("%s", s.Set(128).Eq().Paren())
	// Output: ( ssf = "128" )
}

func ExampleSSF() {
	// convenient alternative to "var X SecurityStrengthFactor, X.Set(...) ..."
	fmt.Printf("%s", SSF(128))
	// Output: 128
}

func ExampleSSF_setLater() {
	s := SSF() // this is functionally the same ...
	// var s SecurityStrengthFactor // ... as this.

	// ... later in your code ...

	fmt.Printf("%s", s.Set(127))
	// Output: 127
}

func TestAnonymous_eqne(t *testing.T) {
	want := `authmethod != "NONE"`
	got := Anonymous.Ne()
	if want != got.String() {
		t.Errorf("%s failed: want '%s', got '%s'", t.Name(), want, got)
	}

	// reset
	want = `authmethod = "NONE"`
	got = Anonymous.Eq()
	if want != got.String() {
		t.Errorf("%s failed: want '%s', got '%s'", t.Name(), want, got)
	}
}

func ExampleAuthMethod_Ne() {
	fmt.Printf("%s", Anonymous.Ne())
	// Output: authmethod != "NONE"
}

func ExampleAuthMethod_Eq() {
	fmt.Printf("%s", SASL.Eq())
	// Output: authmethod = "SASL"
}
