FROM phppm/nginx:latest
LABEL Description="Mikrotik hosts parser application container"

COPY . /app/src

WORKDIR /app/src

RUN \
  echo -e "\n\
alias ls='ls --color=auto';\n\
export PS1='\[\e[1;31m\]\$(echo \"[\"\${?/0/}\"]\" | sed \"s/\\[\\]//\")\$(echo \"\[\e[32m\][hosts-parser] \
\[\e[37m\]\")\u@\h: \[\e[00m\]\w \\$ ';\n\n" >> /root/.bashrc \
  && php --version \
  && export COMPOSER_HOME="/usr/local/share" \
  && export COMPOSER_ALLOW_SUPERUSER="1" \
  && export PATH="$COMPOSER_HOME:$COMPOSER_HOME/vendor/bin:$PATH" \
  && php -r "copy('https://getcomposer.org/installer', '/tmp/composer-setup.php');" \
  && php -r "if(hash_file('SHA384','/tmp/composer-setup.php')==='544e09ee996cdf60ece3804abc52599c22b1f40f4323403c'.\
    '44d44fdfdd586475ca9813a858088ffbc1f233e9b180f061'){echo 'Verified';}else{unlink('/tmp/composer-setup.php');}" \
  && php /tmp/composer-setup.php --filename=composer --install-dir=$COMPOSER_HOME \
  && rm -Rf /tmp/* \
  && $COMPOSER_HOME/composer --no-interaction global require 'hirak/prestissimo' \
  && composer install --no-interaction --no-suggest --no-dev \
  && composer clear-cache \
  && composer dump-autoload \
  && php ./artisan cache:clear

VOLUME ["/app/src"]

CMD ["--bootstrap=laravel", "--app-env=prod", "--workers=8", "--static-directory=public/"]
