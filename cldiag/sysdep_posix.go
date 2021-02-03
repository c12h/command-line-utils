// +build !windows,!plan9

package cldiag

import (
	"fmt"
	"log/syslog"
	"os"
	"strings"
	"syscall"
)

var altDest *os.File

// WriteMessage2 writes an informational message (as opposed to a warning or
// fatal error message), with an optional tag between the prefix and the ":".
//
func WriteMessage2(tag, format string, v ...interface{}) {
	t := new(strings.Builder)
	t.WriteString(msgPrefix)
	if tag == "" {
		t.WriteRune(' ')
		t.WriteString(tag)
	}
	fmt.Fprintf(t, ": "+format+"\n", v...)
	text := t.String()
	if l := len(text); text[l-2] == '\n' {
		text = text[:l-1]
	}
	t = nil

	verb := "write to"
	if altDest == nil { // this is the usual case
		_, err := os.Stderr.WriteString(text)
		if err != nil {
			err = TidyError(err)
		} else {
			err = os.Stderr.Sync()
			if err != nil {
				err = TidyError(err)
				if err == syscall.EINVAL {
					// For un-Sync-able files: pipe, FIFO, sockets, …
					err = nil
				} else {
					verb = "sync"
				}
			}
		}
		if err == nil {
			return // Message successfully written via stderr and synced.
		}

		// Oops! Use /dev/tty instead of os.Stderr for this and future messages.
		var err2 error
		altDest, err2 = os.OpenFile("/dev/tty", os.O_WRONLY, 0666)
		if err2 != nil {
			// “We're all going to die” ... err, panic().
			syslogWriter, err3 :=
				syslog.Dial("", "", syslog.LOG_CRIT, os.Args[0])
			text2 := fmt.Sprintf(
				"can neither %s stderr (%s) nor open /dev/tty (%s)",
				verb, err, err2)
			if err3 != nil {
				// AFAICT, this should not happen.
				panic(fmt.Sprintf("%s PANIC: %s to report: %s",
					os.Args[0], text2, text))
			}
			fmt.Fprintf(syslogWriter, "%s to report: %s",
				text2, text[:len(text)-1])
			panic(fmt.Sprintf("%s PANIC: %s: more in syslog",
				os.Args[0], text2))
		}
		fmt.Fprintf(altDest,
			"%s: cannot %s stderr (%s), using /dev/tty instead\n",
			os.Args[0], verb, err)
	}
	altDest.WriteString(text)
}
