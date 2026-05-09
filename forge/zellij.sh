#!/usr/bin/env bash
####################################################################################################
# zellij
####################################################################################################

# config
. "${HOME}/.lilith/.config.sh"

####################################################################################################

# create temporary files
cat << HEAD >> "${zellij}/.keybinds.tmp"

keybinds clear-defaults=true {

HEAD

cat << HEAD >> "${zellij}/.eof.tmp"

}

HEAD

####################################################################################################

# concatenate
mbombo --in "${zmodes}" --out "${zellij}/config.kdl" \
  --files "../.keybinds.tmp" \
  --files "except.kdl" \
  --files "locked.kdl" \
  --files "shared.kdl" \
  --files "renamepane.kdl" \
  --files "renametab.kdl" \
  --files "entersearch.kdl" \
  --files "search.kdl" \
  --files "../.eof.tmp" \
  --files "header.kdl"

####################################################################################################

# purge temporary files
rm "${zellij}/.keybinds.tmp"
rm "${zellij}/.eof.tmp"

####################################################################################################
