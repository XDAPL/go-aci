package aci

/*
pb.go contains PermissionBindRules types and methods. PermissionBindRules combine
Permissions with BindRules, and are a key component in the formation of a complete
ACI.
*/

/*
invalid value constants used as stringer method returns when
something goes wrong :/
*/
const (
	// badPB is supplied during PermissionBindRule string representation
	// attempts that fail
	badPB = `<invalid_pbrule>`
)

var (
	badPermissionBindRule  PermissionBindRule  // for failed calls that return a PermissionBindRule only
	badPermissionBindRules PermissionBindRules // for failed calls that return a PermissionBindRules only
)

/*
PB contains one (1) Permission instance and one (1) BindRules instance. Instances
of this type are used to create top-level ACI Permission+Bind Rule pairs, of which
at least one (1) is required in any given access control instructor definition.

Users may compose instances of this type manually, or using the PB package
level function, which automatically invokes value checks.
*/
type PermissionBindRule struct {
	P Permission
	B BindContext // BindRule -or- BindRules are allowed
}

/*
PBR returns an instance of PermissionBindRule, bearing the Permission P and
the Bind Rule B. The values P and B shall undergo validity checks per the
conditions of the PermissionBindRule.Valid method automatically.

Instances of this kind are intended for submission (via Push) into instances
of PermissionBindRules.

Generally, an ACI only has a single PermissionBindRule, though multiple
instances of this type are allowed per the syntax specification honored
by this package.
*/
func PBR(P Permission, B BindContext) (r PermissionBindRule) {
	_r := PermissionBindRule{P, B}
	if err := _r.Valid(); err == nil {
		r = _r
	}

	return
}

/*
Valid returns an error instance should any of the following conditions
evaluate as true:

• Permission.Valid returns an error for P

• Rule.Valid returns an error for B

• Rule.Len returns zero (0) for B
*/
func (r *PermissionBindRule) Valid() (err error) {
	return r.valid()
}

func (r *PermissionBindRule) IsZero() bool {
	if r == nil {
		return true
	}

	return r.P.IsZero() && r.B == nil
}

func (r PermissionBindRule) Kind() string {
	return pbrRuleID
}

/*
valid is a private method invoked by PermissionBindRule.Valid.
*/
func (r *PermissionBindRule) valid() (err error) {
	if r == nil {
		err = nilInstanceErr(r)
		return
	}

	if err = r.P.Valid(); err != nil {
		return

	} else if err = r.B.Valid(); err != nil {
		return
	}

	if r.B.Len() == 0 {
		err = nilInstanceErr(r.B)
	} else if r.P.IsZero() {
		err = nilInstanceErr(r.P)
	} else if r.B.ID() != bindRuleID {
		err = badPTBRuleKeywordErr(r.B, bindRuleID, bindRuleID, r.B.ID())
	}

	return
}

/*
String is a stringer method that returns the string representation
of the receiver.
*/
func (r PermissionBindRule) String() string {
	return r.string()
}

/*
Compare returns a Boolean value indicative of a SHA-1 comparison
between the receiver (r) and input value x.
*/
func (r PermissionBindRule) Compare(x any) bool {
	return compareHashInstance(r, x)
}

/*
string is a private method called by PermissionBindRule.String.
*/
func (r PermissionBindRule) string() (s string) {
	s = badPB
	if err := r.valid(); err == nil {
		s = sprintf("%s %s;", r.P, r.B)
	}

	return
}

/*
Valid wraps go-stackage's Stack.Valid method.
*/
func (r PermissionBindRules) Valid() (err error) {
	_t, _ := castAsStack(r)
	err = _t.Valid()
	return
}

/*
IsZero wraps go-stackage's Stack.IsZero method.
*/
func (r PermissionBindRules) IsZero() bool {
	_r, _ := castAsStack(r)
	return _r.IsZero()
}

/*
Len wraps go-stackage's Stack.Len method.
*/
func (r PermissionBindRules) Len() int {
	_r, _ := castAsStack(r)
	return _r.Len()
}

/*
Category wraps go-stackage's Stack.ID method.
*/
func (r PermissionBindRules) Kind() string {
	return pbrsRuleID
}

/*
Index wraps go-stackage's Stack.Index method and performs type
assertion in order to return an instance of PermissionBindRule.
*/
func (r PermissionBindRules) Index(idx int) (pbr PermissionBindRule) {
	_r, _ := castAsStack(r)
	x, _ := _r.Index(idx)

	if assert, asserted := x.(PermissionBindRule); asserted {
		pbr = assert
	}

	return
}

/*
String is a stringer method that returns the string
representation of the receiver instance.

This method wraps go-stackage's Stack.String method.
*/
func (r PermissionBindRules) String() string {
	_r, _ := castAsStack(r)
	return _r.String()
}

/*
Compare returns a Boolean value indicative of a SHA-1 comparison
between the receiver (r) and input value x.
*/
func (r PermissionBindRules) Compare(x any) bool {
	return compareHashInstance(r, x)
}

/*
Push wraps go-stackage's Stack.Push method.
*/
func (r PermissionBindRules) Push(x ...any) PermissionBindRules {
	_r, _ := castAsStack(r)

	// iterate variadic input arguments
	for i := 0; i < len(x); i++ {

		if assert, ok := x[i].(PermissionBindRule); ok {
			// Push it!
			_r.Push(assert)
		}
	}

	return r
}

/*
Pop wraps go-stackage's Stack.Pop method. An instance of
PermissionBindRule, which may or may not be nil, is returned
following a call of this method.

Within the context of the receiver type, a PermissionBindRule,
if non-nil, can only represent a PermissionBindRule instance.
*/
func (r PermissionBindRules) Pop() (pbr PermissionBindRule) {
	_r, _ := castAsStack(r)
	x, _ := _r.Pop()

	if assert, ok := x.(PermissionBindRule); ok {
		pbr = assert
	}

	return
}

/*
permissionBindRulesPushPolicy conforms to the PushPolicy interface signature
defined within the go-stackage package. This private function is called during
Push attempts to a PermissionBindRules instance.
*/
func (r PermissionBindRules) pushPolicy(x any) (err error) {
	if r.contains(x) {
		err = pushErrorNotUnique(r, x, nil)
		return
	}

	switch tv := x.(type) {

	case PermissionBindRule:
		if err = tv.Valid(); err != nil {
			err = pushErrorNilOrZero(PermissionBindRules{}, tv, nil, err)
		}

	default:
		err = pushErrorBadType(PermissionBindRules{}, tv, nil)
	}

	return
}

/*
Contains returns a Boolean value indicative of whether value x,
if a string or PermissionBindRule instance, already resides within
the receiver instance.

Case is not significant in the matching process.
*/
func (r PermissionBindRules) Contains(x any) bool {
	return r.contains(x)
}

/*
contains is a private method called by PermissionBindRules.Contains.
*/
func (r PermissionBindRules) contains(x any) bool {
	if r.Len() == 0 {
		return false
	}

	var candidate string

	switch tv := x.(type) {
	case string:
		candidate = tv
	case PermissionBindRule:
		candidate = tv.String()
	default:
		return false
	}

	if len(candidate) == 0 {
		return false
	}

	for i := 0; i < r.Len(); i++ {
		// case is not significant here.
		if eq(r.Index(i).String(), candidate) {
			return true
		}
	}

	return false
}

/*
PBRs returns a freshly initialized instance of PermissionBindRules, configured to
store one (1) or more instances of PermissionBindRule.

Instances of this kind are used as a component in top-level Instruction (ACI)
assembly.
*/
func PBRs(x ...any) (pbr PermissionBindRules) {
	// create a native stackage.Stack
	// and configure before typecast.
	_p := stackList().
		NoNesting(true).
		SetID(pbrsRuleID).
		SetDelimiter(rune(32)).
		SetCategory(pbrsRuleID).
		NoPadding(!StackPadding)

	// cast _p as a proper PermissionBindRules
	// instance (pbr). We do it this way to gain
	// access to the method for the *specific
	// instance* being created (pbr), thus allowing
	// things like uniqueness checks, etc., to
	// occur during push attempts, providing more
	// helpful and non-generalized feedback.
	pbr = PermissionBindRules(_p)
	_p.SetPushPolicy(pbr.pushPolicy)

	// Assuming one (1) or more items were
	// submitted during the call, (try to)
	// push them into our initialized stack.
	// Note that any failed push(es) will
	// have no impact on the validity of
	// the return instance.
	_p.Push(x...)

	return
}

const pbrRuleID = `pbr`
const pbrsRuleID = `pbrs`
