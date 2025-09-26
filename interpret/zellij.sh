####################################################################################################
# zellij
####################################################################################################

# config
source "${HOME}/.lilith/.workflow_config.sh"

####################################################################################################

# interpret
babel embed --program zellij-entersearch --target "${zmodes}/entersearch.kdl"
babel embed --program zellij-locked      --target "${zmodes}/locked.kdl"
babel embed --program zellij-renamepane  --target "${zmodes}/renamepane.kdl"
babel embed --program zellij-renametab   --target "${zmodes}/renametab.kdl"
babel embed --program zellij-search      --target "${zmodes}/search.kdl"
babel embed --program zellij-shared      --target "${zmodes}/shared.kdl"

####################################################################################################
