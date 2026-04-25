/*
Copyright © 2026 Daniel Rivas <danielrivasmd@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"syscall"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func slayDaemon(daemonMeta string) {
	const op = "lilith.slay"
	meta := loadMeta(daemonMeta)

	sendDaemonSignal(meta.PID, syscall.SIGTERM)

	horus.CheckErr(
		func() error {
			_, err := domovoi.RemoveFile(daemonMeta, rootFlags.verbose)(resolveMetaPath(daemonMeta))
			return err
		}(),
		horus.WithOp(op),
		horus.WithCategory("io_error"),
		horus.WithMessage("removing metadata file"),
	)

	horus.CheckErr(
		func() error {
			_, err := domovoi.RemoveFile(meta.LogPath, rootFlags.verbose)(meta.LogPath)
			return err
		}(),
		horus.WithOp(op),
		horus.WithCategory("io_error"),
		horus.WithMessage("removing log file"),
	)

	fmt.Printf("%s slayed daemon %q\n", chalk.Green.Color("OK:"), meta.Daemon)
}

func slayGroupDaemons(group string) {
	for _, daemonMeta := range listDaemonMetaFiles() {
		if matchDaemonGroup(daemonMeta, group) {
			slayDaemon(daemonMeta)
		}
	}
}

func slayAllDaemons() {
	for _, daemonMeta := range listDaemonMetaFiles() {
		slayDaemon(daemonMeta)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
