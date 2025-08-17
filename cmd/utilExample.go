////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"strings"

	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// formatExample builds a multi‐line example block
// each usage is a slice of “tokens”: [ command, flagOrArg, flagOrArg, ... ].
//
//	app:    your binary name, e.g. "lilith"
//	usages: one or more usages—each becomes its own line.
func formatExample(app string, usages ...[]string) string {
	var b strings.Builder

	for i, usage := range usages {
		if len(usage) == 0 {
			continue
		}

		// first token is the subcommand
		b.WriteString(
			chalk.White.Color(app) + " " +
				chalk.White.Color(chalk.Bold.TextStyle(usage[0])),
		)

		// remaining tokens are either flags (--foo) or args
		for _, tok := range usage[1:] {
			switch {
			case strings.HasPrefix(tok, "--"):
				b.WriteString(" " + chalk.Italic.TextStyle(chalk.White.Color(tok)))
			default:
				b.WriteString(" " + chalk.Dim.TextStyle(chalk.Italic.TextStyle(tok)))
			}
		}

		if i < len(usages)-1 {
			b.WriteRune('\n')
		}
	}

	return b.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var exampleRoot = formatExample(
	"lilith",
	[]string{"help"},
)

var exampleInvoke = formatExample(
	"lilith",
	[]string{"invoke", "--config", "helix"},
	[]string{
		"invoke", "--name", "helix",
		"--watch", "~/src/helix",
		"--script", "helix.sh",
		"--log", "helix",
	},
)

var exampleSlay = formatExample(
	"lilith",
	[]string{"slay", "helix"},
	[]string{"slay", "--group", "<forge>"},
	[]string{"slay", "--all"},
)

var exampleTally = formatExample(
	"lilith",
	[]string{"tally"},
)

var exampleFreeze = formatExample(
	"lilith",
	[]string{"freeze", "helix"},
	[]string{"freeze", "--group", "<forge>"},
	[]string{"freeze", "--all"},
)

var exampleSummon = formatExample(
	"lilith",
	[]string{"summon", "helix", "--follow"},
)

var exampleInstall = formatExample(
	"lilith",
	[]string{"install"},
)

var exampleRekindle = formatExample(
	"lilith",
	[]string{"rekindle", "helix"},
	[]string{"rekindle", "--group", "<forge>"},
	[]string{"rekindle", "--all"},
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
