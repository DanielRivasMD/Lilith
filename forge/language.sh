####################################################################################################
# language
####################################################################################################

# config
source "${HOME}/.lilith/.workflow_config.sh"

####################################################################################################

# concatenate
mbombo --in "${hlangs}" --out "${helix}/languages.toml" \
  --files "awk.toml" \
  --files "bash.toml" \
  --files "csv.toml" \
  --files "docker.toml" \
  --files "clojure.toml" \
  --files "git.toml" \
  --files "go.toml" \
  --files "html.toml" \
  --files "json.toml" \
  --files "julia.toml" \
  --files "kbd.toml" \
  --files "kdl.toml" \
  --files "lua.toml" \
  --files "markdown.toml" \
  --files "nu.toml" \
  --files "python.toml" \
  --files "R.toml" \
  --files "rust.toml" \
  --files "scala.toml" \
  --files "sql.toml" \
  --files "ssh-config.toml" \
  --files "toml.toml" \
  --files "yaml.toml"

####################################################################################################
