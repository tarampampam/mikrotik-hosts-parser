#!/usr/bin/env sh
set -e

PHP_BIN=$(which php);
COMPOSER_BIN=$(which composer);
APP_DIR="${PWD:-/app/src}";

cd "$APP_DIR";

echo '[info] Environment variables:';
printenv;
echo;

"$COMPOSER_BIN" dump-autoload;
"$PHP_BIN" ./artisan cache:clear;

exec "$@"
