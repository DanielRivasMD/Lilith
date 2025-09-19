####################################################################################################
# helix
####################################################################################################

# config
source "${HOME}/.lilith/.workflow_config.sh"

# interpret
babel interpret --program helix-common > "${moded}/common.toml"
babel interpret --program helix-insert > "${moded}/insert.toml"
babel interpret --program helix-normal > "${moded}/normal.toml"
babel interpret --program helix-select > "${moded}/select.toml"

####################################################################################################
