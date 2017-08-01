#!/usr/bin/env bash

export LC_ALL=C;

# Getting directory path with own script
[[ -z $BASE_DIR ]] && export BASE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )";

# Try to find application root directory
if [[ -z $APP_ROOT ]]; then
  if [ -f "$BASE_DIR/../composer.json" ]; then
    export APP_ROOT=$(realpath "$BASE_DIR/../");
  else
    if [ -f "$BASE_DIR/../../composer.json" ]; then
      export APP_ROOT=$(realpath "$BASE_DIR/../../");
    else
      if [ -f "$BASE_DIR/../../../composer.json" ]; then
        export APP_ROOT=$(realpath "$BASE_DIR/../../../");
      fi;
    fi;
  fi;
fi;

if [[ -z $APP_ROOT ]]; then
  echo 'Cannot find application root directory. Exit.' && exit 1;
else
  cd "$APP_ROOT";
fi;

# Declare pathes to binaries
[[ -z $BASH_BIN ]]     && export BASH_BIN=$(which bash);
[[ -z $COMPOSER_BIN ]] && export COMPOSER_BIN=$(which composer);
[[ -z $ARTISAN_BIN ]]  && export ARTISAN_BIN="$APP_ROOT/artisan";
[[ -z $PHP_BIN ]]      && export PHP_BIN=$(which php);
[[ -z $PHPUNIT_BIN ]]  && export PHPUNIT_BIN=$(which phpunit);
[[ -z $GIT_BIN ]]      && export GIT_BIN=$(which git);

# Require another script
function require() {
  source "$1" 2>/dev/null || { echo "[FATAL ERROR] $2 file '$1' not exists, or contains errors" && exit 1; };
}

# Style user text
function style() {
  local user_text=$1;   # User input text (string)
  local user_styles=$2; # Text color/styles, separated by spaces (string)
  declare -A styles;
  styles['white']='\033[0;37m';  # White text color
  styles['red']='\033[0;31m';    # Red text color
  styles['green']='\033[0;32m';  # Green text color
  styles['yellow']='\033[0;33m'; # Yellow text color
  styles['blue']='\033[0;34m';   # Blue text color
  styles['gray']='\033[1;30m';   # Gray text color
  styles['bold']='\033[1m';      # Bold text style
  styles['underline']='\033[4m'; # Underlined text style
  styles['reverse']='\033[7m';   # Reversed colors text style
  styles['none']='\033[0m';      # Reset text styles
  local text_styles='';
  for style in $user_styles; do
    if [[ ! -z "$style" ]] && [[ ! -z "${styles[$style]}" ]]; then
      text_styles="$text_styles${styles[$style]}";
    fi;
  done;
  [ ! -z "$text_styles" ] && {
    echo -e "$text_styles$user_text${styles[none]}";
  } || {
    echo -e "$1";
  };
}

# Show user message
function message() {
  local type=$1;       # Message type (string)
  local text=$2;       # Message text (string)
  local extra=$3;      # Additional text (string)
  local additional=$4; # Additional options
  [ "$#" -eq 1 ] && {
    text=$1;
  };
  [ ! -z "$extra" ] && {
    local styled_text=$(style "$extra" 'yellow');
    text="$text ($styled_text)";
  };
  local text_out='';
  local to_stderr=0;
  local now=$(date +%H:%M:%S);
  case $type in
    'error')   text_out="[$(style $now 'red')] $text"; to_stderr=1;;
    'fatal')   text_out="$(style 'Fatal error:' 'red reverse') $text\n"; to_stderr=1;;
    'notice')  text_out="[$(style $now 'blue bold')] $text";;
    'debug')   text_out="$(style '[Debug   ]' 'reverse') $text";;
    'info')    text_out="[$(style $now 'yellow')] $text";;
    'verbose') text_out="[$(style $now 'yellow')] $text";;
    *)         text_out="[$(date +%H:%M:%S)] $text";;
  esac;
  local additional_flags='';
  # Do not output the trailing newline (no_newline/oneline = no newline)
  if [[ "$additional" == 'no_newline' ]] || [[ "$additional" == 'oneline' ]]; then
    additional_flags="$additional_flags -n";
  fi;
  [ $to_stderr -ne 1 ] && {
    echo -e $additional_flags "$text_out";
  } || {
    echo -e $additional_flags "$text_out" 1>&2;
  };
}

# Copy .env file to the application root directory
function setApplicationEnvFile() {
  local env_file_path="$1"; # Path to the .env file (string)

  if [ -f "$env_file_path" ]; then
    message 'info' "Copy .env file ($env_file_path) to the application root directory ($APP_ROOT)";
    cp -f "$env_file_path" "$APP_ROOT/.env" && return 0;
  else
    message 'error' "Env config ($env_file_path) was not found";
  fi;

  return 1;
}
