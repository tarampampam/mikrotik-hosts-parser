#!/usr/bin/env bash

# Require bootstrap file
source $(dirname "$0")'/bootstrap.sh' || {
  echo '[FATAL ERROR] Bootstrap file not found or contains errors' && exit 1;
};
# Now we can use: $BASE_DIR and $APP_ROOT variables

if [ -d "$APP_ROOT" ]; then
  message 'info' 'Setup directories permissions';
  find "$APP_ROOT" -type d -exec chmod 755 {} \;
  find "$APP_ROOT" -type d -exec chmod ug+s {} \;

  message 'info' 'Setup files permissions';
  find "$APP_ROOT" -type f -exec chmod 644 {} \;
  find "$APP_ROOT" -type f \( \
    -iname "*.sh" \
    -o -name "artisan" \
    -o -name "phpunit" \
    -o -name "*.phar" \
  \) -exec chmod +x {} \;

  message 'info' 'Setup cache directories permissions';
  if [ -d "$APP_ROOT/storage" ]; then
    find "$APP_ROOT/storage" -type d -exec chmod 775 {} \;
    find "$APP_ROOT/storage" -type f -exec chmod 664 {} \;
  fi;
  if [ -d "$APP_ROOT/bootstrap/cache" ]; then
    find "$APP_ROOT/bootstrap/cache" -type d -exec chmod 775 {} \;
    find "$APP_ROOT/bootstrap/cache" -type f -exec chmod 664 {} \;
  fi;

  exit 0;
else
  message 'fatal' "Application root directory ($APP_ROOT) was not found";
fi;

exit 1;
