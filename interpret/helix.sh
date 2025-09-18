####################################################################################################
# helix
####################################################################################################

# config
source "${HOME}/.lilith/.workflow_config.sh"

# interpret
babel interpret --program helix-common > "${dotfiles}/ex-situ/helix/modes/common.toml"
babel interpret --program helix-insert > "${dotfiles}/ex-situ/helix/modes/insert.toml"
babel interpret --program helix-normal > "${dotfiles}/ex-situ/helix/modes/normal.toml"
babel interpret --program helix-select > "${dotfiles}/ex-situ/helix/modes/select.toml"

####################################################################################################
