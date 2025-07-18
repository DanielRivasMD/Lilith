####################################################################################################
# language
####################################################################################################

# config
source "${HOME}/.lilith/sh/config.sh"

# concatenate
mbombo forge --path "${languages}" --out "${helix}/languages.toml" \
  --files "awk.toml" \
  --files "bash.toml" \
  --files "docker.toml" \
  --files "git.toml" \
  --files "go.toml" \
  --files "json.toml" \
  --files "julia.toml" \
  --files "markdown.toml" \
  --files "python.toml" \
  --files "R.toml" \
  --files "rust.toml" \
  --files "scala.toml" \
  --files "sql.toml" \
  --files "toml.toml" \
  --files "yaml.toml"

####################################################################################################

