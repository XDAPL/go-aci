package aci

import (
	"fmt"
	"testing"
)

func TestInheritance(t *testing.T) {
	inh := Inherit(UAT(AT(`manager`), USERDN), Level0, Level1, Level2, Level8)
	want := `userattr = "parent[0,1,2,8].manager#USERDN"`
	got := inh.Eq()
	if want != got.String() {
		t.Errorf("%s failed: want '%s', got '%s'", t.Name(), want, got)
	}
}

func ExampleInherit_uSERDN() {
	attr := AT(`manager`)
	uat := UAT(attr, AV(`USERDN`))
	inh := Inherit(uat, 0, 1, 2, 8)
	fmt.Printf("%s", inh.Eq())
	// Output: userattr = "parent[0,1,2,8].manager#USERDN"
}

func ExampleInherit_uAT() {
	attr := AT(`manager`)
	uat := UAT(attr, AV(`uid=frank,ou=People,dc=example,dc=com`))
	inh := Inherit(uat, 1, 3)
	fmt.Printf("%s", inh.Eq())
	// Output: userattr = "parent[1,3].manager#uid=frank,ou=People,dc=example,dc=com"
}

func ExampleInherit_gAT() {
	attr := AT(`owner`)
	gat := GAT(attr, USERDN)
	inh := Inherit(gat, 3, 4)
	fmt.Printf("%s", inh.Eq())
	// Output: groupattr = "parent[3,4].owner#USERDN"
}

func TestLevels_bogus(t *testing.T) {
	var inh Inheritance
	if err := inh.Valid(); err == nil {
		t.Errorf("%s failed: invalid %T returned no validity error",
			t.Name(), inh)
	}

	if inh.String() != badInheritance {
		t.Errorf("%s failed: invalid %T returned no bogus inheritance warning",
			t.Name(), inh)
	}

	if inh.Eq() != badBindRule {
		t.Errorf("%s failed: invalid %T returned unexpected %T instance during equality bindrule creation",
			t.Name(), inh, badBindRule)
	}

	if inh.Ne() != badBindRule {
		t.Errorf("%s failed: invalid %T returned unexpected %T instance during negated equality bindrule creation",
			t.Name(), inh, badBindRule)
	}

	if !inh.IsZero() {
		t.Errorf("%s failed: bogus %T is non-zero",
			t.Name(), inh)
	}

	for _, rawng := range []string{
		`parent[100].manager#USERDN`,
		`parent[].manager#SELFDN`,
		`parent[4]#ROLEDN`,
		`parent[-1,20,3,476,5,666,7,666,9]?manager#LDAPURI`,
		`parent[0]].owner#GROUPDN`,
		`Parent[1,3,5,7)owner]#LDAPURI`,
		`parent[1,3,5,7)owner#LDAPURI`,
		`parent[1,2,3,4].squatcobbler`,
		``,
	} {
		i, err := parseInheritance(rawng)
		if err == nil {
			t.Errorf("%s failed: parsing of bogus %T definition returned no error",
				t.Name(), i)

		}

		if i.String() != badInheritance {
			t.Errorf("%s failed: %T parsing attempt failed; want '%s', got '%s'",
				t.Name(), i, badInheritance, i)
		}
	}
}

func TestInheritance_parse(t *testing.T) {
	for _, raw := range []string{
		`parent[0,5,9].manager#USERDN`,
		`parent[1].manager#SELFDN`,
		`parent[4].terminated#ROLEDN`,
		`parent[0,1,2,3,4,5,6,7,8,9].manager#LDAPURI`,
		`parent[0].owner#GROUPDN`,
	} {
		i, err := parseInheritance(raw)
		if err != nil {
			t.Errorf("%s failed: %T parsing attempt failed; %v",
				t.Name(), i, err)

		}

		if raw != i.String() {
			t.Errorf("%s failed: %T parsing attempt failed; want '%s', got '%s'",
				t.Name(), i, raw, i)
		}

		want := sprintf("( userattr = %q )", raw)
		equality := i.Eq().Paren()

		if got := equality.String(); want != got {
			t.Errorf("%s failed: %T equality creation error; want '%s', got '%s'",
				t.Name(), i, want, got)
		}

		negation := i.Ne().Paren()
		want = sprintf("( userattr != %q )", raw)
		if got := negation.String(); want != got {
			t.Errorf("%s failed: %T negated equality creation error; want '%s', got '%s'",
				t.Name(), i, want, got)
		}
	}
}

/*
func TestInheritance_lrShift(t *testing.T) {
        var p Inheritance

        // three iterations, one per supported
        // Level type
        for i := 0; i < 3; i++ {

                // iterate each of the levels in the
                // levels/names map
                for k, v := range levelMap {

                        term, typ := testGetLevelsTermType(i, k, v.String())

                        shifters := map[int]func(...any) Inheritance{
                                0: p.Shift,
                                1: p.Unshift,
                        }

                        for j := 0; j < len(shifters); j++ {
                                mode, phase := testGetLevelsPhase(j)
                                if shifters[j](term); p.Positive(term) != phase {
                                        t.Errorf("%s failed: %T %s %s failed [key:%d; term:%v] (value:%v)",
                                                t.Name(), p, typ, mode, k, term, p)
                                }
                        }
                }
        }
}

func testGetLevelsPhase(j int) (mode string, phase bool) {
        mode = `shift`
        if phase = (0 == j); !phase {
                mode = `un` + mode
        }

        return
}

func testGetLevelsTermType(i,k int, v string) (term any, typ string) {
        term = k // default
        switch i {
        case 1:
                term = v // string name (e.g.: `0`)
        case 2:
                term = Level(k) // Level
        }
        typ = sprintf("%T", term) // label for err

        return
}
*/
