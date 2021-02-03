// This package writes error messages in the conventional format
//	<Program-Name>: <Sprintf-output><newline>
// or the slightly less conventional format
//	<Program> <Tag>: <Sprintf-output><newline>
// where the tag can be a string such as "BUG", "summary", etc.
//
// All messages go to standard error (at least at first; if a write to os.Stderr
// fails, this package tries system-dependent alternatives).
//
// Most calls into this package will be to two sets of functions, called
// Warn[If][2] and Die[If][2] hereinafter.  Here's a summary of these functions:
//	* Warn(format, formatArgs)
//	* Warn2(tag, format, formatArgs)
//	* WarnIf(skipIfNil, format, formatArgs)
//	* WarnIf2(skipIfNil, tag, format, formatArgs)
//	* Die(format, formatArgs)
//	* Die2(tag, format, formatArgs)
//	* DieIf(skipIfNil, format, formatArgs)
//	* DieIf2(skipIfNil, tag, format, formatArgs)
//   where
//	- format is a format string for use with fmt.Sprintf()
//	- fmtArgs are arguments for that format string
//	- tag is a string to go after the program name in the above-mentioned
//	  ‘slightly less conventional format’, or "" to use the more
//	  conventional format.
//	- if skipIfNil is nil, WarnIf[2] and DieIf[2] do nothing.
//   There are some special cases:
//	- WarnIf[2](err, "") is equivalent to WarnIf[2](err, "%s", err)
//	- DieIf[2](err, "")  is equivalent to DieIf[2](err, "%s", err)
//	- Die("") and Die2(tag, "") call os.Exit() without outputting any message.
//   This package provides three levels of diagnostic:
//	- the WriteMessage[2] functions are for informational messages
//	- the Warn[If][2] functions count how many warnings are output
//	- the Die[If][2] functions call os.Exit()
//
// When calling os.Exit(), the default exit status is 3 if any warnings were
// reported, or 2 if none were.  Programs can call SetExitStatus to change this.
//
// This module has a subpackage named cldiag_no_prefix which provides wrappers
// for the Warn[If][2] and Die[If][2] functions. It is intended to be imported
// without a prefix, like this:
//	import . "github.com/c12h/command-line-utils/cldiag/cldiag_no_prefix"
//	...
//		err = open_backup(...)
//		DieIf(err, "cannot restore from backup: %s", err)
// The idea is that the names of these functions are distinctive enough to not need
// prefixing.
//
// Dying is Dangerous
//
// WARNING: calling os.Exit() will NOT run deferred functions in the current
// goroutine, let alone in other goroutines. Therefore, it is strongly
// recommended that you only call Die[If][2] (1) before doing anything that
// might create another goroutine or defer any cleanup operations or (2) in
// main() after getting an error that needs to be reported as fatal. To be
// explicit: only main(), closely-related functions and setup code should ever
// call these routines.
//
package cldiag // import "github.com/c12h/command-line-utils/cldiag"

// ???FIXME: This should go in a _example file.
//	if nFatal > 0 {
//		cldiag.Die("%d fatal syntactic error(s) found", nFatal)
//	}

import (
	"fmt"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"strings"
)

/*============== Package-level variables and related functions ===============*/

var (
	msgPrefix string = filepath.Base(os.Args[0])
	nWarnings int    = 0
)

// NumberOfWarnings returns the number of times a Warn[If][2] call has written a
// warning. (WarnIf[2] calls with SkipIfNil==nil don’t count.)
//
func NumberOfWarnings() int {
	return nWarnings
}

// GetPrefix returns the string that goes at the start of every diagnostic
// written by these function.
//
// This string is set to filepath.Base(os.Args[0]) on program start.
func GetPrefix() string {
	return msgPrefix
}

// SetPrefix changes the string that these functions put at the start of every
// message.
//
// For example, some programs might use
//		cldiag.SetMessagePrefix(os.Args[0])
// to avoid ambiguity.
func SetPrefix(s string) {
	msgPrefix = s
}

/*------------------------------- Exit status --------------------------------*/

var exitStatus int = 2

func dieExitStatus() int {
	if nWarnings > 0 {
		return exitStatus | 1
	}
	return exitStatus
}

// GetExitStatus returns the base exit status that would be used if a Die[If][2]
// function reported a fatal error.
//
func GetExitStatus() int { return dieExitStatus() }

// SetExitStatus sets a new base exit status.  It panics on values < 2 or > 124.
// The new value should usually be an even number.
//
// Any Die[If][2] routine which calls os.Exit() will use the given value, except
// that if any warning have been reported, they use (value | 1).
func SetExitStatus(newStatus int) int {
	oldStatus := exitStatus
	if newStatus < 2 || newStatus > 124 {
		Panic("SetExitStatus(%d): need 2 to 124 inclusive", newStatus)
	}
	exitStatus = newStatus
	return oldStatus
}

/*=========================== Writing Diagnostics ============================*/

/*-------------------------- Informational Messages --------------------------*/

// WriteMessage writes an informational message (as opposed to a warning or
// fatal error message).
func WriteMessage(format string, fmtArgs ...interface{}) {
	WriteMessage2("", format, fmtArgs...)
}

// WriteMessage2() is system-dependent

/*--------------------------------- Warnings ---------------------------------*/

// Warn writes a warning message.
func Warn(format string, fmtArgs ...interface{}) {
	Warn2("", format, fmtArgs...)
}

// Warn2 writes a warning message.  It takes an optional ‘tag’ argument.
func Warn2(tag, format string, fmtArgs ...interface{}) {
	nWarnings++
	WriteMessage2(tag, format, fmtArgs...)
}

// WarnIf writes a warning message if (and only if) its first argument is non-nil.
//
// As a special case, WarnIf(x,"") is equivalent to WarnIf(x,"%s",x).
func WarnIf(skipIfNil interface{}, format string, fmtArgs ...interface{}) {
	WarnIf2(skipIfNil, "", format, fmtArgs...)
}

// WarnIf writes a warning message if (and only if) its first argument is non-nil.
// It takes an optional ‘tag’ argument.
//
// As a special case, WarnIf2(x,tag,"") is equivalent to WarnIf2(x,tag,"%s",x).
func WarnIf2(skipIfNil interface{}, tag, format string, fmtArgs ...interface{}) {
	if skipIfNil != nil {
		if format == "" {
			Warn2(tag, "%s", skipIfNil)
		} else {
			Warn2(tag, format, fmtArgs...)
		}
	}
}

/*------------------------------- Fatal Errors -------------------------------*/

// Die writes a fatal error message and calls os.Exit.  As a special case, it
// does not write an error message if the format string is empty.
func Die(format string, fmtArgs ...interface{}) {
	Die2("", format, fmtArgs...)
}

// Die2 writes a fatal error message (if and only if format is non-empty) and
// calls os.Exit.  It takes an optional ‘tag’ argument.
func Die2(tag, format string, fmtArgs ...interface{}) {
	if format != "" {
		WriteMessage2(tag, format, fmtArgs...)
	}
	//
	os.Exit(dieExitStatus())
}

// DieIf reports a fatal error and calls os.Exit().
//
// As a special case, DieIf(x,"") is equivalent to DieIf(x,"%s",x).
func DieIf(skipIfNil interface{}, format string, fmtArgs ...interface{}) {
	if skipIfNil == nil {
		return
	} else if format == "" {
		Die2("", "%s", skipIfNil)
	} else {
		Die2("", format, fmtArgs...)
	}
}

// DieIf2 reports a fatal error and calls os.Exit().
// It takes an optional ‘tag’ argument.
//
// As a special case, DieIf2(x,tag,"") is equivalent to DieIf2(x,tag,"%s",x).
func DieIf2(skipIfNil interface{}, tag, format string, fmtArgs ...interface{}) {
	if skipIfNil == nil {
		return
	} else if format == "" {
		Die2(tag, "%s", skipIfNil)
	} else {
		Die2(tag, format, fmtArgs...)
	}
}

/*---------------------------------- Panics ----------------------------------*/

// Panic is a wrapper for the built-in panic() function which produces messages
// in the same format as Warn(), Die() etc.
func Panic(format string, fmtArgs ...interface{}) {
	text := msgPrefix + ": " + fmt.Sprintf(format, fmtArgs...)
	panic(text)
}

// Panic2 is a wrapper for the built-in panic() function which produces messages
// in the same format as Warn2(), Die2() etc (ie, it takes an optional ‘tag’
// argument).
func Panic2(tag, format string, fmtArgs ...interface{}) {
	if tag != "" {
		tag = " " + tag
	}
	text := msgPrefix + tag + ": " + fmt.Sprintf(format, fmtArgs...)
	panic(text)
}

// TidyError(e) is equivalent to e.Error() except for a few special cases,
// in which it gives something users of command-line programs will (IMO) find easier
// to understand.
//
// Currently, TidyError() unwraps (pointers to) os.PathError, os.LinkError and
// os.SyscallError values.  More cases may be added in the future.
func TidyError(e error) error {
	switch ee := e.(type) {
	case *os.PathError:
		// os.PathError.Error() produces texts of the form
		//	<Operation> <Path>: <BaseError>
		// (without any quote characters!). This package assumes you
		// provide the first two items in your main error text, so
		// you'll want the base error’s text instead of the PathError’s.
		return ee.Unwrap()
	case *os.LinkError:
		// os.LinkError.Error() produces
		//	<Operation> <Path1> <Path2>: <BaseError>

		// again without quotes, which is not optimal if either path
		// contains spaces.  Assuming you provide the first three items
		// in a wrapper error, you'll want the base error’s text instead
		// of the LinkError’s.
		return ee.Unwrap()
	case *os.SyscallError:
		// os.SyscallError.Error() produces just
		//	<SyscallName>: <BaseError>
		// which is often too cryptic for non-programmers.  You usually
		// should provide a user-friendly verb to the syscall and report
		// the base error’s text instead of the SyscallError’s.
		return ee.Unwrap()
	case *urlpkg.Error:
		// Errors from the net.url package can generate texts of the form
		//	net/url: <problem in url>
		// which
		if ee.Op == "parse" {
			text := ee.Unwrap().Error()
			if strings.HasPrefix(text, "net/url: ") {
				return &UntidyError{
					OriginalError: e,
					BaseError:     ee,
					TrimStart:     len("net/url: ")}
			}
		}
	default:
		text := e.Error()
		// If only the author(s) of archive/zip had defined a custom error type,
		// we could look at type info instead of playing dubious tricks with the
		// generated string.
		const zipPrefix = "zip: "
		if text[:len(zipPrefix)] == zipPrefix {
			return &UntidyError{
				OriginalError: e,
				BaseError:     ee,
				TrimStart:     len(zipPrefix)}
		}
	}
	return e
}

type UntidyError struct {
	OriginalError error
	BaseError     error
	TrimStart     int
}

func (e *UntidyError) Error() string { return e.BaseError.Error()[e.TrimStart:] }
func (e *UntidyError) Unwrap() error { return e.BaseError }
