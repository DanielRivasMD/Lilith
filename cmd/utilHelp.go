////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// formatHelp produces the “help” header + description.
//
//	author: name, e.g. "Daniel Rivas"
//	email:  email, e.g. "danielrivasmd@gmail.com"
//	desc:   the multi‐line description, "\n"-separated.
func formatHelp(author, email, desc string) string {
	header := chalk.Bold.TextStyle(
		chalk.Green.Color(author+" "),
	) +
		chalk.Dim.TextStyle(
			chalk.Italic.TextStyle("<"+email+">"),
		)

	// prefix two newlines to your desc, chalk it cyan + dim it
	body := "\n\n" + desc
	return header + chalk.Dim.TextStyle(chalk.Cyan.Color(body))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpRoot = formatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Master of daemons",
)

var helpInvoke = formatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Spawn daemon process for the specified directory & execute the configured script on change\n"+
		"Metadata is persistent for summoning the daemon",
)

var helpSlay = formatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Gracefully stop alive daemons, removing their metadata and logs to allow clean reinvocation",
)

var helpTally = formatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"List all daemons invoked, showing group, PID, start time, and current status",
)

var helpFreeze = formatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Pause daemon execution using SIGSTOP, until resumed manually",
)

var helpSummon = formatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Display daemon log output\n"+
		"Pass "+chalk.Italic.TextStyle("--follow")+" to stream in real time",
)

var helpInstall = formatHelp(
	"Daniel Rivas ",
	"<danielrivasmd@gmail.com>",
	"Install lilith config",
)

var helpRekindle = formatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Restart daemons in limbo using persisted metadata",
)

////////////////////////////////////////////////////////////////////////////////////////////////////
