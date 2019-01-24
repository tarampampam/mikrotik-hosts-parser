FROM composer:1.8.0 AS composer

FROM phppm/nginx:latest
LABEL Description="Mikrotik hosts parser application container"

ENV COMPOSER_ALLOW_SUPERUSER="1" \
    COMPOSER_HOME="/tmp/composer" \
    PS1='\[\033[1;32m\]🐳 \[\033[1;36m\][\u@\h] \[\033[1;34m\]\w\[\033[0;35m\] \[\033[1;36m\]# \[\033[0m\]'

COPY --from=composer /usr/bin/composer /usr/bin/composer
COPY . /app/src

WORKDIR /app/src

RUN set -xe \
    && php --version \
    && rm -Rf /tmp/* \
    && composer global require 'hirak/prestissimo' --no-interaction --no-suggest --prefer-dist \
    && composer install --no-dev --no-interaction --no-ansi --no-suggest --prefer-dist \
    && composer clear-cache \
    && composer dump-autoload \
    && php ./artisan cache:clear

VOLUME ["/app/src"]

CMD ["--bootstrap=laravel", "--app-env=prod", "--workers=4", "--static-directory=public/"]
