# MikroTik hosts parser

[![Version][badge_packagist_version]][link_packagist]
[![Version][badge_php_version]][link_packagist]
[![PPM Compatible][badge_ppm]][link_ppm]
[![Build Status][badge_build_status]][link_build_status]
[![Coverage][badge_coverage]][link_coverage]
[![Code quality][badge_code_quality]][link_coverage]
[![Downloads count][badge_downloads_count]][link_packagist]
[![License][badge_license]][link_license]

Приложение, которое генерирует скрипт для роутера на базе `RouterOS`, который блокирует "рекламные" хосты.

Более подробно о нем можно прочитать по [этой ссылке (хабр)][habr].

### Установка

Для развертывания приложения достаточно выполнить в терминале:

```bash
$ composer create-project tarampampam/mikrotik-hosts-parser
```

Все интересные настройки вынесены в файлы конфигурации, что лежат в директории `./config`:

### Docker

[![Docker build][badge_docker_build]][link_docker_build]
[![Docker pulls][badge_docker_pulls]][link_docker_pulls]
[![Docker size][badge_docker_size]][link_docker_pulls]

[/r/tarampampam/mikrotik-hosts-parser][docker_hub]

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

## Changes log

[![Release date][badge_release_date]][link_releases]
[![Commits since latest release][badge_commits_since_release]][link_commits]

Changes log can be [found here][link_changes_log].

## Support

[![Issues][badge_issues]][link_issues]
[![Issues][badge_pulls]][link_pulls]

If you will find any package errors, please, [make an issue][link_create_issue] in current repository.

## License

This is open-sourced software licensed under the [MIT License][link_license].

[badge_packagist_version]:https://img.shields.io/packagist/v/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&maxAge=180
[badge_php_version]:https://img.shields.io/packagist/php-v/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&longCache=true
[badge_ppm]:https://raw.githubusercontent.com/php-pm/ppm-badge/master/ppm-badge.png
[badge_build_status]:https://img.shields.io/scrutinizer/build/g/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&maxAge=180&logo=scrutinizer
[badge_code_quality]:https://img.shields.io/scrutinizer/g/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&maxAge=180
[badge_coverage]:https://img.shields.io/scrutinizer/coverage/g/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&maxAge=180
[badge_downloads_count]:https://img.shields.io/packagist/dt/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&maxAge=180
[badge_license]:https://img.shields.io/packagist/l/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&longCache=true
[badge_release_date]:https://img.shields.io/github/release-date/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&maxAge=180
[badge_commits_since_release]:https://img.shields.io/github/commits-since/tarampampam/mikrotik-hosts-parser/latest.svg?style=flat-square&maxAge=180
[badge_issues]:https://img.shields.io/github/issues/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&maxAge=180
[badge_pulls]:https://img.shields.io/github/issues-pr/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&maxAge=180
[badge_docker_build]:https://img.shields.io/docker/build/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&maxAge=180
[badge_docker_pulls]:https://img.shields.io/docker/pulls/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&maxAge=180
[badge_docker_size]:https://images.microbadger.com/badges/image/tarampampam/mikrotik-hosts-parser:latest.svg?style=flat-square
[link_releases]:https://github.com/tarampampam/mikrotik-hosts-parser/releases
[link_packagist]:https://packagist.org/packages/tarampampam/mikrotik-hosts-parser
[link_ppm]:https://github.com/php-pm/php-pm
[link_build_status]:https://scrutinizer-ci.com/g/tarampampam/mikrotik-hosts-parser/build-status/master
[link_coverage]:https://scrutinizer-ci.com/g/tarampampam/mikrotik-hosts-parser/?branch=master
[link_changes_log]:https://github.com/tarampampam/mikrotik-hosts-parser/blob/master/CHANGELOG.md
[link_issues]:https://github.com/tarampampam/mikrotik-hosts-parser/issues
[link_create_issue]:https://github.com/tarampampam/mikrotik-hosts-parser/issues/new/choose
[link_commits]:https://github.com/tarampampam/mikrotik-hosts-parser/commits
[link_pulls]:https://github.com/tarampampam/mikrotik-hosts-parser/pulls
[link_license]:https://github.com/tarampampam/mikrotik-hosts-parser/blob/master/LICENSE
[link_docker_build]:https://hub.docker.com/r/tarampampam/mikrotik-hosts-parser/builds/
[link_docker_pulls]:https://hub.docker.com/r/tarampampam/mikrotik-hosts-parser/
[getcomposer]:https://getcomposer.org/download/
[demo]: https://stopad.hook.sh/
[habr]: https://habrahabr.ru/post/264001/
[docker_hub]:https://hub.docker.com/r/tarampampam/mikrotik-hosts-parser/
