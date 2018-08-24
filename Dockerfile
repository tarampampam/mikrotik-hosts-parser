FROM php:7.2-alpine
LABEL Description="Mikrotik hosts parser application container"

ENV COMPOSER_HOME /usr/local/share/composer
ENV COMPOSER_ALLOW_SUPERUSER 1
ENV PATH "$COMPOSER_HOME:$COMPOSER_HOME/vendor/bin:$PATH"

# Install basic deps
RUN \
  docker-php-ext-install opcache \
  && php --version \
  && mkdir -pv $COMPOSER_HOME && chmod -R g+w $COMPOSER_HOME \
  && php -r "copy('https://getcomposer.org/installer', '/tmp/composer-setup.php');" \
  && php -r "if(hash_file('SHA384','/tmp/composer-setup.php')==='544e09ee996cdf60ece3804abc52599c22b1f40f4323403c'.\
    '44d44fdfdd586475ca9813a858088ffbc1f233e9b180f061'){echo 'Verified';}else{unlink('/tmp/composer-setup.php');}" \
  && php /tmp/composer-setup.php --filename=composer --install-dir=$COMPOSER_HOME \
  && $COMPOSER_HOME/composer --no-interaction global require 'hirak/prestissimo' \
  && $COMPOSER_HOME/composer --version && $COMPOSER_HOME/composer global info \
  && rm -rf /tmp/composer-setup* \
  && mkdir -pv /app/src

WORKDIR /app/src

# Copy application sources and configs
COPY . /app/src
COPY ./docker-entrypoint.sh /docker-entrypoint.sh

# Make composer install and configure other applications
RUN \
  cd /app/src \
  && composer install --no-interaction --no-suggest --no-dev \
  && composer clear-cache \
  && chmod +x /docker-entrypoint.sh \
  && php ./artisan

STOPSIGNAL SIGTERM
EXPOSE 80
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["php", "-S", "0.0.0.0:80", "-t", "./public"]
