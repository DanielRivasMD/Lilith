####################################################################################################
# helix
####################################################################################################

# config
source "${HOME}/.lilith/.workflow_config.sh"

####################################################################################################

# interpret
babel interpret --program helix-common > "${hmodes}/common.toml"
babel interpret --program helix-insert > "${hmodes}/insert.toml"
babel interpret --program helix-normal > "${hmodes}/normal.toml"
babel interpret --program helix-select > "${hmodes}/select.toml"

####################################################################################################
