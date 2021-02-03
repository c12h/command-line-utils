// Package no_prefix is a subpackage of cmdiag which is intended to imported
// without a prefix, like so:
//	import . "github.com/c12h/command-line-utils/cldiag/cldiag_no_prefix"
//
// It provides wrappers for the Die[If][2] and Warn[If][2] functions from
// github.com/c12h/cmdUtils whose names are, I think, distinctive enough to not
// need prefixing.
//
package cldiag_no_prefix

import (
	"github.com/c12h/command-line-utils/cldiag"
)

func Warn(format string, fmtArgs ...interface{}) {
	cldiag.Warn2("", format, fmtArgs)
}
func Warn2(tag, format string, fmtArgs ...interface{}) {
	cldiag.Warn2(tag, format, fmtArgs...)
}

func WarnIf(skipIfNil interface{}, format string, fmtArgs ...interface{}) {
	cldiag.WarnIf2(skipIfNil, "", format, fmtArgs)
}
func WarnIf2(skipIfNil interface{}, tag, format string, fmtArgs ...interface{}) {
	cldiag.WarnIf2(skipIfNil, tag, format, fmtArgs)
}

func Die(format string, fmtArgs ...interface{}) {
	cldiag.Die2("", format, fmtArgs...)
}
func Die2(tag, format string, fmtArgs ...interface{}) {
	cldiag.Die2(tag, format, fmtArgs)
}

func DieIf(skipIfNil interface{}, format string, fmtArgs ...interface{}) {
	cldiag.DieIf2(skipIfNil, "", format, fmtArgs)
}

func DieIf2(skipIfNil interface{}, tag, format string, fmtArgs ...interface{}) {
	cldiag.DieIf2(skipIfNil, tag, format, fmtArgs)
}
