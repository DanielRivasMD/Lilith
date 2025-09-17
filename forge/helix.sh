####################################################################################################
# helix
####################################################################################################

# config
source "${HOME}/.lilith/.config.sh"

# create temporary files
for type in normal insert select
do
  mbombo forge --in "${moded}" --out "${helix}/.${type}.tmp" --files "common.toml" --replace MODE="${type}_mode"
done

# concatenate
mbombo forge --in "${moded}" --out "${helix}/config.toml" \
  --files "theme.toml" \
  --files "editor.toml" \
  --files "mini-mode.toml" \
  --files "normal.toml" \
  --files "../.normal.tmp" \
  --files "insert.toml" \
  --files "../.insert.tmp" \
  --files "select.toml" \
  --files "../.select.tmp"

# purge temporary files
rm "${helix}/.normal.tmp"
rm "${helix}/.insert.tmp"
rm "${helix}/.select.tmp"

####################################################################################################
