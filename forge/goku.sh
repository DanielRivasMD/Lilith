#!/usr/bin/env bash
####################################################################################################
# goku
####################################################################################################

# config
. "${HOME}/.lilith/.config.sh"

# relocate
oldd=$(pwd)
cd "${saiyajin}"

# generate edn files
find "$ssrc" -type f -name "*.clj" | while read -r filepath; do
  parent=$(dirname "$filepath")

  # skip top-level files
  if [ "$parent" = "$ssrc" ]; then
    continue
  fi

  # skip anything under config/
  case "$parent" in
    */config) continue ;;
    */config/*) continue ;;
  esac

  filename=$(basename "$filepath" .clj)
  subdir=$(basename "$parent")
  clojure -M -m "${subdir}.${filename}"
done

# relocate
cd "${oldd}"

####################################################################################################

# create temporary files
cat << HEAD >> "${sedn}/.header.tmp"
{:profiles
HEAD

cat << HEAD >> "${sedn}/.main.tmp"
:main [
HEAD

cat << HEAD >> "${sedn}/.eof.tmp"
  ]}]}
HEAD

####################################################################################################

# concatenate
mbombo --in "${sedn}" --out "${karabiner}/karabiner.edn" \
  --files ".header.tmp" \
  --files "header.edn" \
  --files ".main.tmp" \
  --files "browser.edn" \
  --files "finder.edn" \
  --files "zoom.edn" \
  --files "q.edn" \
  --files "w.edn" \
  --files "z.edn" \
  --files "tab.edn" \
  --files "joker.edn" \
  --files "lcmd.edn" \
  --files "lctl.edn" \
  --files "lopt.edn" \
  --files "lshift.edn" \
  --files "rcmd.edn" \
  --files "rctl.edn" \
  --files "rshift.edn" \
  --files "ropt.edn" \
  --files "keyboards.edn" \
  --files "keymod.edn" \
  --files ".eof.tmp"

####################################################################################################

# purge temporary files
rm "${sedn}/.header.tmp"
rm "${sedn}/.main.tmp"
rm "${sedn}/.eof.tmp"

# render config
goku

####################################################################################################
