////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"github.com/DanielRivasMD/domovoi"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpRoot = domovoi.FormatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Master of daemons",
)

var helpInvoke = domovoi.FormatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Spawn daemon process for the specified directory & execute the configured script on change\n"+
		"Metadata is persistent for summoning the daemon",
)

var helpSlay = domovoi.FormatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Gracefully stop alive daemons, removing their metadata and log to allow clean reinvocation",
)

var helpTally = domovoi.FormatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"List all daemons invoked, showing group, PID, start time, and current status",
)

var helpFreeze = domovoi.FormatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Pause daemon execution using SIGSTOP, until resumed manually",
)

var helpSummon = domovoi.FormatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Display daemon log output\n"+
		"Pass "+chalk.Italic.TextStyle("--follow")+" to stream in real time",
)

var helpRekindle = domovoi.FormatHelp(
	"Daniel Rivas",
	"danielrivasmd@gmail.com",
	"Restart daemons in limbo using persisted metadata",
)

var helpGenesis = domovoi.FormatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"Install config directories & Generate a commented example of a Lilith TOML config",
)

////////////////////////////////////////////////////////////////////////////////////////////////////
