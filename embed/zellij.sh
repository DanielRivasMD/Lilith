####################################################################################################
# zellij
####################################################################################################

# config
source "${HOME}/.lilith/.config.sh"

####################################################################################################

# interpret
babel embed --program zellij-enter  --target "${zmodes}/entersearch.kdl"
babel embed --program zellij-lock   --target "${zmodes}/locked.kdl"
babel embed --program zellij-rpane  --target "${zmodes}/renamepane.kdl"
babel embed --program zellij-rtab   --target "${zmodes}/renametab.kdl"
babel embed --program zellij-search --target "${zmodes}/search.kdl"
babel embed --program zellij-shared --target "${zmodes}/shared.kdl"

####################################################################################################
