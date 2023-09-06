package aci

import (
	"fmt"
	"testing"
)

func TestACIs(t *testing.T) {
	var Ins Instructions
	_ = Ins.Valid()
	_ = Ins.IsZero()
	_ = Ins.String()
	_ = Ins.Len()
	_ = Ins.Index(0)

	// Make a target rule that encompasses any account
	// with a DN syntax of "uid=<userid>,ou=People,dc=example,dc=com"
	C := TDN(`uid=*,ou=People,dc=example,dc=com`).Eq()

	// push into a new instance of Rule automatically
	// configured to store Target Rule Condition instances.
	tgt := TRs(C)

	// define a timeframe for our PermissionBindRule
	// using two Condition instances
	notBefore := ToD(`1730`).Ge()                    // Condition: greater than or equal to time
	notAfter := ToD(`2400`).Lt()                     // Condition: less than time
	brule := And().Paren().Push(notBefore, notAfter) // our actual bind rule expression

	// Define the permission (rights).
	perms := Allow(ReadAccess, CompareAccess, SearchAccess)

	// Make our PermissionBindRule instance, which defines the
	// granting of access within a particular timeframe.
	pbrule := PBR(perms, brule)

	// The ACI's effective name (should be unique within the directory)
	acl := `Limit people access to timeframe`

	// Finally, craft the Instruction instance
	var i Instruction
	_ = i.TRs()
	_ = i.PBRs()
	_ = i.ACL()
	_ = i.Valid()
	_ = i.String()

	i = ACI(acl, tgt, pbrule)
	_ = i.TRs()
	_ = i.PBRs()
	_ = i.ACL()

	want := `( target = "ldap:///uid=*,ou=People,dc=example,dc=com" )(version 3.0; acl "Limit people access to timeframe"; allow(read,search,compare) ( timeofday >= "1730" AND timeofday < "2400" );)`
	if want != i.String() {
		t.Errorf("%s failed: want '%s', got '%s'", t.Name(), want, i)
		return
	}

	Ins = ACIs()
	Ins.Push(i)
	popped := Ins.Pop()
	Ins.Push(popped)
	Ins.F()
	Ins.Push(popped.String())
	Ins.Push(`<3 <3 <3`)
	if Ins.Len() != 1 {
		t.Errorf("%s failed to push %T into %T, len:%d, want:%d\n%s", t.Name(), i, Ins, Ins.Len(), 1, Ins)
		return
	}

	if Ins.String() != Ins.Index(0).String() {
		t.Errorf("%s strcmp fail", t.Name())
		return
	}
}

func ExampleInstruction_build() {
	// Make a target rule that encompasses any account
	// with a DN syntax of "uid=<userid>,ou=People,dc=example,dc=com"
	C := TDN(`uid=*,ou=People,dc=example,dc=com`).Eq()

	// push into a new instance of Rule automatically
	// configured to store Target Rule Condition instances.
	tgt := TRs(C).Push(C)

	// define a timeframe for our PermissionBindRule
	// using two Condition instances
	notBefore := ToD(`1730`).Ge()                    // Condition: greater than or equal to time
	notAfter := ToD(`2400`).Lt()                     // Condition: less than time
	brule := And().Paren().Push(notBefore, notAfter) // our actual bind rule expression

	// Define the permission (rights).
	perms := Allow(ReadAccess, CompareAccess, SearchAccess)

	// Make our PermissionBindRule instance, which defines the
	// granting of access within a particular timeframe.
	pbrule := PBR(perms, brule)

	// The ACI's effective name (should be unique within the directory)
	acl := `Limit people access to timeframe`

	// Finally, craft the Instruction instance
	var i Instruction
	i.Set(acl, tgt, pbrule)

	fmt.Printf("%s", i)
	// Output: ( target = "ldap:///uid=*,ou=People,dc=example,dc=com" )(version 3.0; acl "Limit people access to timeframe"; allow(read,search,compare) ( timeofday >= "1730" AND timeofday < "2400" );)
}

func ExampleInstruction_buildNested() {
	// Make a target rule that encompasses any account
	// with a DN syntax of "uid=<userid>,ou=People,dc=example,dc=com"
	C := TDN(`uid=*,ou=People,dc=example,dc=com`).Eq()

	// push into a new instance of Rule automatically
	// configured to store Target Rule Condition instances.
	tgt := TRs().Push(C)

	// create an ORed stack, pushing the two specified
	// userdn equality conditions into its collection.
	ors := Or().Paren().Push(
		UDN(`uid=jesse,ou=admin,dc=example,dc=com`).Eq(),
		UDN(`uid=courtney,ou=admin,dc=example,dc=com`).Eq(),
	)

	// create our AttributeBindTypeOrValue instance,
	// setting the AttributeType as `ninja`, and the
	// AttributeValue as `FALSE`
	attr := AT(`ninja`)       // attributeType
	aval := AV(`FALSE`)       // attributeValue
	userat := UAT(attr, aval) // attributeBindTypeOrValue

	// create a negated (NOT) stack, pushing
	// our AttributeBindTypeOrValue BindRule
	// (Eq()) instance into its collection.
	// Also, make stack parenthetical.
	nots := Not().Paren().Push(userat.Eq())

	// define a timeframe for our PermissionBindRule
	// using two Condition instances. Make both AND
	// stacks parenthetical, and push our OR and NOT
	// stacks defined above.
	brule := And().Paren().Push(
		And().Paren().Push(
			ToD(`1730`).Ge(), // Condition: greater than or equal to time
			ToD(`2400`).Lt(), // Condition: less than time
		),
		ors,
		nots,
	)

	// Define the permission (rights).
	perms := Allow(ReadAccess, CompareAccess, SearchAccess)

	// Make our PermissionBindRule instance, which defines the
	// granting of access within a particular timeframe.
	pbr := PBR(perms, brule)

	// The ACI's effective name (should be unique within the directory)
	acl := `Limit people access to timeframe`

	// Finally, craft the Instruction instance
	var i Instruction
	i.Set(acl, tgt, pbr)

	fmt.Printf("%s", i)
	// Output: ( target = "ldap:///uid=*,ou=People,dc=example,dc=com" )(version 3.0; acl "Limit people access to timeframe"; allow(read,search,compare) ( ( timeofday >= "1730" AND timeofday < "2400" ) AND ( userdn = "ldap:///uid=jesse,ou=admin,dc=example,dc=com" OR userdn = "ldap:///uid=courtney,ou=admin,dc=example,dc=com" ) AND NOT ( userattr = "ninja#FALSE" ) );)
}

/*
This example demonstrates doing a literal search for an ACI within a stack of ACIs
using its Contains method. Case is not significant in the matching process.
*/
func ExampleInstructions_Contains() {
	raw1 := `( target = "ldap:///uid=*,ou=People,dc=example,dc=com" )(version 3.0; acl "Limit people access to timeframe for those ninjas"; allow(read,search,compare) ( ( timeofday >= "1730" AND timeofday < "2400" ) AND ( userdn = "ldap:///uid=jesse,ou=admin,dc=example,dc=com" OR userdn = "ldap:///uid=courtney,ou=admin,dc=example,dc=com" ) AND NOT ( userattr = "ninja#FALSE" ) );)`
	raw2 := `( target = "ldap:///uid=*,ou=People,dc=example,dc=com" )(version 3.0; acl "Limit people access to timeframe"; allow(read,search,compare) ( timeofday >= "1730" AND timeofday < "2400" );)`

	acis := ACIs(
		raw1,
		raw2,
	)

	fmt.Printf("%T contains raw1: %t", acis, acis.Contains(raw1))
	// Output: aci.Instructions contains raw1: true
}

/*
This example demonstrates use of the F method to obtain the
package-level function appropriate for the creation of new
stack elements.
*/
func ExampleInstructions_F() {
	var acis Instructions
	funk := acis.F()

	ins := funk() // normally you'd want to supply some type instances
	fmt.Printf("%T", ins)
	// Output: aci.Instruction

}

func ExampleACIs() {
	raw1 := `( target = "ldap:///uid=*,ou=People,dc=example,dc=com" )(version 3.0; acl "Limit people access to timeframe for those ninjas"; allow(read,search,compare) ( ( timeofday >= "1730" AND timeofday < "2400" ) AND ( userdn = "ldap:///uid=jesse,ou=admin,dc=example,dc=com" OR userdn = "ldap:///uid=courtney,ou=admin,dc=example,dc=com" ) AND NOT ( userattr = "ninja#FALSE" ) );)`
	raw2 := `( target = "ldap:///uid=*,ou=People,dc=example,dc=com" )(version 3.0; acl "Limit people access to timeframe"; allow(read,search,compare) ( timeofday >= "1730" AND timeofday < "2400" );)`

	acis := ACIs(
		raw1,
		raw2,
	)

	fmt.Printf("%T contains %d Instruction instances", acis, acis.Len())
	// Output: aci.Instructions contains 2 Instruction instances
}
