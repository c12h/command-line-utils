// Package clerrs provides a generic error type, and some related utility
// functions.
//
// Its generic errors which are (pointers to) clerrs.CannotError, a struct type
// suitable for English-language error messages of the form
//	cannot <verb>[ <adjective>]  <o-q-noun>[ <suffix>[: <base-error>]
// where <o-q-noun> is a noun phrase that can be put in double quotes (using the
//  "fmt" package’s %q verb).
//
// Package clerrs also provides some utility functions for making simple HTTP
// GET and HEAD requests and dealing with any resulting errors.
//
package clerrs // import "github.com/c12h/command-line-utils/cldiag"

import (
	"fmt"
	"strings"

	"github.com/c12h/command-line-utils/cldiag"
)

// A CannotError holds details of a problem suitable for messages of the form
//	cannot <verb>[ <adjective>]  <o-q-noun>[ <suffix>[: <base-error>]
// which assumes English (sorry!).
//
type CannotError struct {
	Verb      string // A present-tense verb
	Adjective string // What kind of thing the action was on, or ""
	Noun      string // Which thing the action was on
	QuoteNoun bool   // Whether to put .Noun in double quotes
	Suffix    string // Text to go after the noun, or ""
	BaseError error  // The underlying error, if any
}

// Cannot() is a convenience function to produce a (pointer to a) CannotError value.
//
// Many callers will want to define
//	var cannot = clerrs.Cannot
// to save precious columns when writing error messages.
//
func Cannot(
	verb, adjective, noun string,
	quoteNoun bool, suffix string, baseError error,
) *CannotError {
	return &CannotError{
		Verb:      verb,
		Adjective: adjective,
		Noun:      noun,
		QuoteNoun: quoteNoun,
		Suffix:    suffix,
		BaseError: baseError,
	}
}

// Pointers to CannotError values satisfy the error interface.
func (ce *CannotError) Error() string {
	var b strings.Builder
	b.WriteString("cannot " + ce.Verb + " ")
	if ce.Adjective != "" {
		b.WriteString(ce.Adjective + " ")
	}
	if ce.QuoteNoun {
		fmt.Fprintf(&b, "%q", ce.Noun)
	} else {
		b.WriteString(ce.Noun)
	}
	if ce.Suffix != "" {
		b.WriteString(" " + ce.Suffix)
	}
	if ce.BaseError != nil {
		b.WriteString(": ")
		b.WriteString(cldiag.TidyError(ce.BaseError).Error())
	}
	return b.String()
}

// A CannotError may specify an underlying error.
func (ce *CannotError) Unwrap() error {
	return ce.BaseError
}
