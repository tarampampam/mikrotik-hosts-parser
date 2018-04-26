FROM php:7.2
LABEL Description="Mikrotik hosts parser application container"

# Install basic deps
RUN \
  apt-get -yq update && apt-get -yq upgrade -o Dpkg::Options::="--force-confold" \
  && apt-get -yq install --no-install-recommends openssl unzip zip git \
  && docker-php-ext-install opcache \
  && php --version

# Install composer
ENV COMPOSER_HOME /usr/local/share/composer
ENV COMPOSER_ALLOW_SUPERUSER 1
ENV PATH "$COMPOSER_HOME:$COMPOSER_HOME/vendor/bin:$PATH"
RUN \
  mkdir -pv $COMPOSER_HOME && chmod -R g+w $COMPOSER_HOME \
  && curl -o /tmp/composer-setup.php https://getcomposer.org/installer \
  && curl -o /tmp/composer-setup.sig https://composer.github.io/installer.sig \
  && php -r "if (hash('SHA384', file_get_contents('/tmp/composer-setup.php')) \
    !== trim(file_get_contents('/tmp/composer-setup.sig'))) { unlink('/tmp/composer-setup.php'); \
    echo 'Invalid installer' . PHP_EOL; exit(1); }" \
  && php /tmp/composer-setup.php --filename=composer --install-dir=$COMPOSER_HOME \
  && $COMPOSER_HOME/composer --no-interaction global require 'hirak/prestissimo' \
  && $COMPOSER_HOME/composer --version && $COMPOSER_HOME/composer global info \
  && rm -rf /tmp/composer-setup*

# Copy application sources and configs
RUN mkdir -pv /app/src
COPY . /app/src
COPY ./docker-entrypoint.sh /docker-entrypoint.sh

# Make composer install and configure other applications
RUN \
  cd /app/src \
  && composer install --no-interaction --no-dev \
  && chmod +x /docker-entrypoint.sh \
  && php ./artisan

# Make clear
RUN apt-get -yqq clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /app/src
STOPSIGNAL SIGTERM
EXPOSE 80
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["php", "-S", "0.0.0.0:80", "-t", "./public"]
