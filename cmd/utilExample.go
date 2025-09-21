////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"github.com/DanielRivasMD/domovoi"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var exampleRoot = domovoi.FormatExample(
	"lilith",
	[]string{"help"},
)

var exampleInvoke = domovoi.FormatExample(
	"lilith",
	[]string{"invoke", "helix"},
	[]string{
		"invoke", "--daemon", "helix",
		"--watch", "~/src/helix",
		"--script", "helix.sh",
		"--log", "helix",
	},
)

var exampleSlay = domovoi.FormatExample(
	"lilith",
	[]string{"slay", "helix"},
	[]string{"slay", "--group", "<forge>"},
	[]string{"slay", "--all"},
)

var exampleTally = domovoi.FormatExample(
	"lilith",
	[]string{"tally"},
)

var exampleFreeze = domovoi.FormatExample(
	"lilith",
	[]string{"freeze", "helix"},
	[]string{"freeze", "--group", "<forge>"},
	[]string{"freeze", "--all"},
)

var exampleSummon = domovoi.FormatExample(
	"lilith",
	[]string{"summon", "helix", "--follow"},
)

var exampleRekindle = domovoi.FormatExample(
	"lilith",
	[]string{"rekindle", "helix"},
	[]string{"rekindle", "--group", "<forge>"},
	[]string{"rekindle", "--all"},
)

var exampleGenesis = domovoi.FormatExample(
	"lilith",
	[]string{"genesis", "--output", "/Users/drivas/.lilith/config/example.toml"},
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
