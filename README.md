# MikroTik hosts parser

[![Version][badge_version]][link_packagist]
[![Docker build][badge_docker_build]][link_docker_build]
[![Build Status][badge_build_status]][link_build_status]
[![StyleCI][badge_styleci]][link_styleci]
[![Coverage][badge_coverage]][link_coverage]
[![License][badge_license]][link_license]
[![Downloads count][badge_downloads_count]][link_packagist]
[![Docker pulls][badge_docker_pulls]][link_docker_pulls]

Приложение, которое генерирует скрипт для роутера на базе `RouterOS`, который блокирует "рекламные" хосты.

Более подробно о нем можно прочитать по [этой ссылке (хабр)][habr].

### Установка

Для развертывания приложения достаточно выполнить в терминале:

```bash
$ composer create-project tarampampam/mikrotik-hosts-parser
```

Все интересные настройки вынесены в файлы конфигурации, что лежат в директории `./config`:

### Docker

[/r/avtodev/docker-php71-pg-redis][docker_hub]

Для pull-а образа контейнера:

```bash
$ docker pull tarampampam/mikrotik-hosts-parser
```

Для "быстрого" запуска:

```bash
$ docker run --rm -p 8000:80 tarampampam/mikrotik-hosts-parser
```

И откройте в вашем браузере `http://127.0.0.1:8000`

### Демо

Не гарантирую что приложение будет жить вечно, что пользоваться им можешь [тут][demo].

[badge_version]:http://img.shields.io/packagist/v/tarampampam/mikrotik-hosts-parser.svg?style=flat&maxAge=30
[badge_docker_build]:https://img.shields.io/docker/build/tarampampam/mikrotik-hosts-parser.svg?style=flat&maxAge=30
[badge_docker_pulls]:https://img.shields.io/docker/pulls/tarampampam/mikrotik-hosts-parser.svg?style=flat&maxAge=30
[badge_downloads_count]:https://img.shields.io/packagist/dt/tarampampam/mikrotik-hosts-parser.svg?style=flat&maxAge=30
[badge_license]:https://img.shields.io/packagist/l/tarampampam/mikrotik-hosts-parser.svg
[badge_build_status]:https://scrutinizer-ci.com/g/tarampampam/mikrotik-hosts-parser/badges/build.png?b=master
[badge_styleci]:https://styleci.io/repos/39877790/shield?style=flat&maxAge=30
[badge_coverage]:https://scrutinizer-ci.com/g/tarampampam/mikrotik-hosts-parser/badges/coverage.png?b=master
[link_packagist]:https://packagist.org/packages/tarampampam/mikrotik-hosts-parser
[link_docker_build]:https://hub.docker.com/r/tarampampam/mikrotik-hosts-parser/builds/
[link_docker_pulls]:https://hub.docker.com/r/tarampampam/mikrotik-hosts-parser/
[link_styleci]:https://styleci.io/repos/39877790/
[link_license]:https://github.com/tarampampam/mikrotik-hosts-parser/blob/master/LICENSE
[link_build_status]:https://scrutinizer-ci.com/g/tarampampam/mikrotik-hosts-parser/build-status/master
[link_coverage]:https://scrutinizer-ci.com/g/tarampampam/mikrotik-hosts-parser/?branch=master
[faker_repository_link]:https://github.com/fzaninotto/Faker
[getcomposer]:https://getcomposer.org/download/
[demo]: https://stopad.kplus.pro/
[habr]: https://habrahabr.ru/post/264001/
[docker_hub]:https://hub.docker.com/r/tarampampam/mikrotik-hosts-parser/
