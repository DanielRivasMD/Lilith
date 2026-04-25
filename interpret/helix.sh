#!/usr/bin/env bash
####################################################################################################
# helix
####################################################################################################

# config
. "${HOME}/.lilith/.config.sh"

####################################################################################################

# interpret
babel interpret --program helix-common --target "${hmodes}/common.toml"
babel interpret --program helix-insert --target "${hmodes}/insert.toml"
babel interpret --program helix-normal --target "${hmodes}/normal.toml"
babel interpret --program helix-select --target "${hmodes}/select.toml"

####################################################################################################
