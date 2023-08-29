package aci

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/JesseCoretta/go-stackage"
)

// AttributeFilterOperationsCommaDelim represents the default
// delimitation scheme offered by this package. In cases where
// the AttributeFilterOperations type is used to represent any
// TargetRule bearing the targattrfilters keyword context, two
// (2) different delimitation characters may be permitted for
// use (depending on the product in question).
//
// Use of this constant allows the use of a comma (ASCII #44) to
// delimit the slices in an AttributeFilterOperations instance as
// opposed to the alternative delimiter (semicolon, ASCII #59).
//
// This constant may be fed to the SetDelimiter method that is
// extended through the AttributeFilterOperations type.
const AttributeFilterOperationsCommaDelim = 0

// AttributeFilterOperationsSemiDelim represents an alternative
// delimitation scheme offered by this package. In cases where
// the AttributeFilterOperations type is used to represent any
// TargetRule bearing the targattrfilters keyword context, two
// (2) different delimitation characters may be permitted for
// use (depending on the product in question).
//
// Use of this constant allows the use of a semicolon (ASCII #59)
// to delimit the slices in an AttributeFilterOperations instance
// as opposed to the default delimiter (comma, ASCII #44).
//
// This constant may be fed to the SetDelimiter method that is
// extended through the AttributeFilterOperations type.
const AttributeFilterOperationsSemiDelim = 1

// MultivalOuterQuotes represents the default quotation style
// used by this package. In cases where a multi-valued BindRule
// or TargetRule expression involving LDAP distinguished names,
// ASN.1 Object Identifiers (in dot notation) and LDAP Attribute
// Type names is being created, this constant will enforce only
// outer-most double-quotation of the whole sequence of values.
//
// Example: keyword = "<val> || <val> || <val>"
//
// This constant may be fed to the SetQuoteStyle method that is
// extended through eligible types.
const MultivalOuterQuotes = 0

// MultivalSliceQuotes represents an alternative quotation scheme
// offered by this package. In cases where a multi-valued BindRule
// or TargetRule expression involving LDAP distinguished names,
// ASN.1 Object Identifiers (in dot notation) and LDAP Attribute
// Type names is being created, this constant shall disable outer
// most quotation and will, instead, quote individual values. This
// will NOT enclose symbolic OR (||) delimiters within quotations.
//
// Example: keyword = "<val>" || "<val>" || "<val>"
//
// This constant may be fed to the SetQuoteStyle method that is
// extended through eligible types.
const MultivalSliceQuotes = 1

/*
ComparisonOperator constants defined within the stackage package are aliased
within this package for convenience, without the need for user-invoked stackage
package import.
*/
const (
	badCop stackage.ComparisonOperator = stackage.ComparisonOperator(0x0)

	Eq stackage.ComparisonOperator = stackage.Eq // 0x1, "Equal To"
	Ne stackage.ComparisonOperator = stackage.Ne // 0x2, "Not Equal to"     !! USE WITH CAUTION !!
	Lt stackage.ComparisonOperator = stackage.Lt // 0x3, "Less Than"
	Le stackage.ComparisonOperator = stackage.Le // 0x4, "Less Than Or Equal"
	Gt stackage.ComparisonOperator = stackage.Gt // 0x5, "Greater Than"
	Ge stackage.ComparisonOperator = stackage.Ge // 0x6, "Greater Than Or Equal"
)

var (
	comparisonOperatorMap              map[string]stackage.ComparisonOperator
	permittedTargetComparisonOperators map[TargetKeyword][]stackage.ComparisonOperator
	permittedBindComparisonOperators   map[BindKeyword][]stackage.ComparisonOperator
)

/*
matchCOP reads the *string representation* of a
stackage.ComparisonOperator instance and returns
the appropriate stackage.ComparisonOperator const.

A bogus stackage.ComparisonOperator (badCop, 0x0)
shall be returned if a match was not made.
*/
func matchCOP(op string) stackage.ComparisonOperator {
	for k, v := range comparisonOperatorMap {
		if op == k {
			return v
		}
	}

	return badCop
}

/*
keywordAllowsComparisonOperator returns a boolean value indicative of
whether Keyword input value kw allows stackage.ComparisonOperator op
for use in T/B rule instances.

Certain keywords, such as TargetScope, allow only certain operators,
while others, such as BindSSF, allow the use of ALL operators.
*/
func keywordAllowsComparisonOperator(kw, op any) bool {
	// identify the comparison operator,
	// save as cop var.
	var cop stackage.ComparisonOperator
	switch tv := op.(type) {
	case string:
		cop = matchCOP(tv)
	case stackage.ComparisonOperator:
		cop = tv
	default:
		return false
	}

	// identify the keyword, and
	// pass it onto the appropriate
	// map search function.
	switch tv := kw.(type) {
	case string:
		if bkw := matchBKW(tv); bkw != BindKeyword(0x0) {
			return bindKeywordAllowsComparisonOperator(bkw, cop)

		} else if tkw := matchTKW(tv); tkw != TargetKeyword(0x0) {
			return targetKeywordAllowsComparisonOperator(tkw, cop)
		}
	case BindKeyword:
		return bindKeywordAllowsComparisonOperator(tv, cop)
	case TargetKeyword:
		return targetKeywordAllowsComparisonOperator(tv, cop)
	}

	return false
}

/*
bindKeywordAllowsComparisonOperator is a private function called by keywordAllowsCompariso9nOperator.
*/
func bindKeywordAllowsComparisonOperator(key BindKeyword, cop stackage.ComparisonOperator) bool {
	// look-up the keyword within the permitted cop
	// map; if found, obtain slices of cops allowed
	// by said keyword.
	cops, found := permittedBindComparisonOperators[key]
	if !found {
		return false
	}

	// iterate the cops slice, attempting to perform
	// a match of the input cop candidate value and
	// the current cops slice [i].
	for i := 0; i < len(cops); i++ {
		if cop == cops[i] {
			return true
		}
	}

	return false
}

/*
targetKeywordAllowsComparisonOperator is a private function called by keywordAllowsCompariso9nOperator.
*/
func targetKeywordAllowsComparisonOperator(key TargetKeyword, cop stackage.ComparisonOperator) bool {
	// look-up the keyword within the permitted cop
	// map; if found, obtain slices of cops allowed
	// by said keyword.
	cops, found := permittedTargetComparisonOperators[key]
	if !found {
		return false
	}

	// iterate the cops slice, attempting to perform
	// a match of the input cop candidate value and
	// the current cops slice [i].
	for i := 0; i < len(cops); i++ {
		if cop == cops[i] {
			return true
		}
	}

	return false
}

/*
RulePadding is a global variable that will be applies to ALL
TargetRule and Bindrule instances assembled during package operations.
This is a convenient alternative to manually invoking the NoPadding
method on a case-by-case basis.

Padding is enabled by default, and can be disabled here globally,
or overridden for individual TargetRule/BindRule instances as needed.

Note that altering this value will not impact instances that were
already created; this only impacts the creation of new instances.
*/
var RulePadding bool = true

/*
StackPadding is a global variable that will be applies to ALL Stack
instances assembled during package operations. This is a convenient
alternative to manually invoking the NoPadding method on a case by
case basis.

Padding is enabled by default, and can be disabled here globally,
or overridden for individual Stack instances as needed.

Note that altering this value will not impact instances that were
already created; this only impacts the creation of new instances.
*/
var StackPadding bool = true

/*
frequently-accessed import function aliases.
*/
var (
	lc       func(string) string                 = strings.ToLower
	uc       func(string) string                 = strings.ToUpper
	eq       func(string, string) bool           = strings.EqualFold
	ctstr    func(string, string) int            = strings.Count
	idxf     func(string, func(rune) bool) int   = strings.IndexFunc
	idxr     func(string, rune) int              = strings.IndexRune
	idxs     func(string, string) int            = strings.Index
	hasPfx   func(string, string) bool           = strings.HasPrefix
	hasSfx   func(string, string) bool           = strings.HasSuffix
	repAll   func(string, string, string) string = strings.ReplaceAll
	contains func(string, string) bool           = strings.Contains
	split    func(string, string) []string       = strings.Split
	trimS    func(string) string                 = strings.TrimSpace
	trimPfx  func(string, string) string         = strings.TrimPrefix
	join     func([]string, string) string       = strings.Join
	printf   func(string, ...any) (int, error)   = fmt.Printf
	sprintf  func(string, ...any) string         = fmt.Sprintf
	atoi     func(string) (int, error)           = strconv.Atoi
	isDigit  func(rune) bool                     = unicode.IsDigit
	isLetter func(rune) bool                     = unicode.IsLetter
	isLower  func(rune) bool                     = unicode.IsLower
	isUpper  func(rune) bool                     = unicode.IsUpper
	uint16g  func([]byte) uint16                 = binary.BigEndian.Uint16
	uint16p  func([]byte, uint16)                = binary.BigEndian.PutUint16
	valOf    func(x any) reflect.Value           = reflect.ValueOf
	typOf    func(x any) reflect.Type            = reflect.TypeOf

	stackOr   func(...int) stackage.Stack = stackage.Or
	stackAnd  func(...int) stackage.Stack = stackage.And
	stackNot  func(...int) stackage.Stack = stackage.Not
	stackList func(...int) stackage.Stack = stackage.List
)

func isAlnum(r rune) bool {
	return isLower(r) || isUpper(r) || isDigit(r)
}

/*
isIdentifier scans the input string val and judges whether
it appears to qualify as an identifier, in that:

- it begins with a lower alpha
- it contains only alphanumeric characters, hyphens or semicolons

This is used, specifically, it identify an LDAP attributeType (with
or without a tag), or an LDAP matchingRule.
*/
func isIdentifier(val string) bool {
	if len(val) == 0 {
		return false
	}

	// must begin with lower alpha.
	if !isLower(rune(val[0])) {
		return false
	}

	// can only end in alnum.
	if !isAlnum(rune(val[len(val)-1])) {
		return false
	}

	for i := 0; i < len(val); i++ {
		ch := rune(val[i])
		switch {
		case isAlnum(ch):
			// ok
		case ch == ';', ch == '-':
			// ok
		default:
			return false
		}
	}

	return true
}

/*
strInSlice returns a boolean value indicative of whether the
specified string (str) is present within slice. Please note
that case is a significant element in the matching process.
*/
func strInSlice(str string, slice []string) bool {
	for i := 0; i < len(slice); i++ {
		if str == slice[i] {
			return true
		}
	}
	return false
}

func isPowerOfTwo(x int) bool {
	return x&(x-1) == 0
}

/*
keywordFromCategory attempts to locate the Category
method from input value r and, if found, runs it.

If a value is obtained, a resolution is attempted in
order to identify it as a BindKeyword or TargetKeyword
instance, which is then returned. Nil is returned in
all other cases.

DECOM me (someday)
*/
func keywordFromCategory(r any) Keyword {
	if r == nil {
		return nil
	}

	// if the instance has the Category
	// func, use reflect to get it.
	meth := getCategoryFunc(r)
	if meth == nil {
		return nil
	}

	var kw any

	// Try to match the category as a target rule
	// keyword context ...
	if tk := matchTKW(meth()); tk != TargetKeyword(0x0) {
		kw = tk
		return kw.(TargetKeyword)

		// Try to match the category as a bind rule
		// keyword context ...
	} else if bk := matchBKW(meth()); bk != BindKeyword(0x0) {
		kw = bk
		return kw.(BindKeyword)
	}

	return nil
}

/*
stackByDNKeyword returns an instance of BindDistinguishedNames based on the
following:

• BindGDN (groupdn) keyword returns the BindDistinguishedNames instance created by GDNs()

• BindRDN (roledn) keyword returns the BindDistinguishedNames instance created by RDNs()

• BindUDN (userdn) keyword returns the BindDistinguishedNames instance created by UDNs()

This function is private and is used only during parsing of bind and target rules
which permit a list of DNs as a single logical value. It exists mainly to keep
cyclomatics low.
*/
func stackByBDNKeyword(key Keyword) BindDistinguishedNames {
	// prepare a stack for our DN value(s)
	// based on the input keyword (key)
	switch key {
	case BindRDN:
		return RDNs()
	case BindGDN:
		return GDNs()
	}

	return UDNs()
}

/*
stackByDNKeyword returns an instance of TargetDistinguishedNames based on the
following:

• Target (target) keyword returns the TargetDistinguishedNames instance created by TDNs()

• TargetTo (target_to) keyword returns the TargetDistinguishedNames instance created by TTDNs()

• TargetFrom (target_from) keyword returns the TargetDistinguishedNames instance created by TFDNs()

This function is private and is used only during parsing of bind and target rules
which permit a list of DNs as a single logical value. It exists mainly to keep
cyclomatics low.
*/
func stackByTDNKeyword(key Keyword) TargetDistinguishedNames {
	// prepare a stack for our DN value(s)
	// based on the input keyword (key)
	switch key {
	case TargetTo:
		return TTDNs()
	case TargetFrom:
		return TFDNs()
	}

	return TDNs()
}

/*
stackByOIDKeyword returns an instance of ObjectIdentifiers based on the following:

• TargetExtOp (extop) keyword returns the ObjectIdentifiers instance created by ExtOps()

• TargetCtrl (targetcontrol) keyword returns the ObjectIdentifiers instance created by Ctrls()
*/
func stackByOIDKeyword(key Keyword) ObjectIdentifiers {
	// prepare a stack for our OID value(s)
	// based on the input keyword (key)
	switch key {
	case TargetExtOp:
		return ExtOps()
	}

	return Ctrls()
}

/*
castAsCondition merely wraps (casts, converts) and returns an
instance of BindRule -OR- TargetRule as a stackage.Condition
instance. This is useful for calling methods that have not been
extended (wrapped) in this package via go-stackage, as it may not
be needed in many cases ...

An instance submitted as x that is neither a BindRule or TargetRule
will result in an empty stackage.Condition return value.

Note this won't alter an existing BindRule or TargetRule instance,
rather a new reference is made through the stackage.Condition type
defined within go-stackage. The BindRule or TargetRule, once it has
been altered to one's satisfaction, can be sent off as intended and
this "Condition Counterpart" can be discarded, or left for GC.
*/
func castAsCondition(x any) (c *stackage.Condition) {
	switch tv := x.(type) {

	// case match is a single BindRule instance
	case BindRule:
		C := stackage.Condition(tv)
		return &C

	// case match is a single TargetRule instance
	case TargetRule:
		C := stackage.Condition(tv)
		return &C
	}

	return nil
}

/*
castAsStack merely wraps (casts, converts) and returns any type
alias of stackage.Stack as a native stackage.Stack.

This is useful for calling methods that have not been extended
(wrapped) in this package via go-stackage, as it might not be
needed in most cases ...

An instance submitted as x that is NOT a type alias of stackage.Stack
will result in an empty stackage.Stack return value.

Note this won't alter an existing values, rather a new reference is
made through the stackage.Condition type defined within go-stackage.
The alias type, once it has been altered to one's satisfaction, can be
sent off as intended and this "Stack Counterpart" can be discarded, or
left for GC.
*/
func castAsStack(u any) (S stackage.Stack, converted bool) {
	switch tv := u.(type) {

	case ObjectIdentifiers:
		converted = true
		S = stackage.Stack(tv)

	case BindDistinguishedNames,
		TargetDistinguishedNames:
		S, converted = castDNRules(tv)

	case BindRules, TargetRules,
		PermissionBindRules:
		S, converted = castBTRules(tv)

	case AttributeTypes:
		converted = true
		S = stackage.Stack(tv)

	case AttributeFilterOperation,
		AttributeFilterOperations:
		S, converted = castFilterRules(tv)
	}

	return
}

func castBTRules(x any) (S stackage.Stack, converted bool) {
	switch tv := x.(type) {
	case BindRules:
		S = stackage.Stack(tv)
		converted = true
	case TargetRules:
		S = stackage.Stack(tv)
		converted = true
	case PermissionBindRules:
		S = stackage.Stack(tv)
		converted = true
	}

	return
}

func castDNRules(x any) (S stackage.Stack, converted bool) {
	switch tv := x.(type) {
	case BindDistinguishedNames:
		S = stackage.Stack(tv)
		converted = true
	case TargetDistinguishedNames:
		S = stackage.Stack(tv)
		converted = true
	}

	return
}

func castFilterRules(x any) (S stackage.Stack, converted bool) {
	switch tv := x.(type) {
	case AttributeFilterOperation:
		S = stackage.Stack(tv)
		converted = true
	case AttributeFilterOperations:
		S = stackage.Stack(tv)
		converted = true
	}

	return
}

/*
getCategoryFunc uses reflect to obtain and return a given
type instance's Category method, if present. If not, nil
is returned.
*/
func getCategoryFunc(x any) func() string {
	v := valOf(x)
	if v.IsZero() {
		return nil
	}

	method := v.MethodByName(`Category`)
	if method.Kind() == reflect.Invalid {
		return nil
	}

	if meth, ok := method.Interface().(func() string); ok {
		return meth
	}

	return nil
}

/*
BindContext is a convenient interface type that is qualified by
the following types:

• BindRule

• BindRules

The qualifying methods shown below are intended to make the
handling of a structure of (likely nested) BindRules instances
slightly easier without an absolute need for type assertion at
every step. These methods are inherently read-only in nature
and represent only a subset of the available methods exported
by the underlying qualifier types.

To alter the underlying value, or to gain access to all of a
given type's methods, type assertion shall be necessary.
*/
type BindContext interface {
	// String returns the string representation of the
	// receiver instance.
	String() string

	// Keyword returns the BindKeyword, enveloped as a
	// Keyword interface value. If the receiver is an
	// instance of BindRule, the value is derived from
	// the Keyword method. If the receiver is an instance
	// of BindRules, the value is derived (and resolved)
	// using the Category method.
	Keyword() Keyword

	// IsZero returns a Boolean value indicative of the
	// receiver instance being nil, or unset.
	IsZero() bool

	// Len returns the integer length of the receiver.
	// Only meaningful when run on BindRules instances.
	Len() int

	// IsNesting returns a Boolean value indicative of
	// whether the receiver contains a stack as a value.
	// Only meaningful when run on BindRules instances.
	IsNesting() bool

	// Category will report `bind` in all scenarios.
	Category() string

	// Kind will report `stack` for a BindRules instance, or
	// `condition` for a BindRule instance
	Kind() string

	// isBindContextQualifier ensures no greedy interface
	// matching outside of the realm of bind rules. It need
	// not be accessed by users, nor is it run at any time.
	isBindContextQualifier() bool
}

func init() {
	comparisonOperatorMap = map[string]stackage.ComparisonOperator{
		Eq.String(): Eq,
		Ne.String(): Ne,
		Lt.String(): Lt,
		Le.String(): Le,
		Gt.String(): Gt,
		Ge.String(): Ge,
	}

	// populate the allowed comparison operator map per each
	// possible TargetRule keyword
	permittedTargetComparisonOperators = map[TargetKeyword][]stackage.ComparisonOperator{
		Target:            {Eq, Ne},
		TargetTo:          {Eq, Ne},
		TargetFrom:        {Eq, Ne},
		TargetCtrl:        {Eq, Ne},
		TargetAttr:        {Eq, Ne},
		TargetExtOp:       {Eq, Ne},
		TargetScope:       {Eq},
		TargetFilter:      {Eq, Ne},
		TargetAttrFilters: {Eq},
	}

	// populate the allowed comparison operator map per each
	// possible BindRule keyword
	permittedBindComparisonOperators = map[BindKeyword][]stackage.ComparisonOperator{
		BindUDN: {Eq, Ne},
		BindRDN: {Eq, Ne},
		BindGDN: {Eq, Ne},
		BindIP:  {Eq, Ne},
		BindAM:  {Eq, Ne},
		BindDNS: {Eq, Ne},
		BindUAT: {Eq, Ne},
		BindGAT: {Eq, Ne},
		BindDoW: {Eq, Ne},
		BindSSF: {Eq, Ne, Lt, Le, Gt, Ge},
		BindToD: {Eq, Ne, Lt, Le, Gt, Ge},
	}
}
