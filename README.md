# :fire: PHP version of this application is no longer supported. This branch is read-only

# MikroTik hosts parser

[![Version][badge_packagist_version]][link_packagist]
[![Version][badge_php_version]][link_packagist]
[![PPM Compatible][badge_ppm]][link_ppm]
[![License][badge_license]][link_license]

Приложение, которое генерирует скрипт для роутера на базе `RouterOS`, который блокирует "рекламные" хосты.

Более подробно о нем можно прочитать по [этой ссылке (хабр)][habr].

### Установка

Для развертывания приложения достаточно выполнить в терминале:

```bash
$ composer create-project tarampampam/mikrotik-hosts-parser
```

Все интересные настройки вынесены в файлы конфигурации, что лежат в директории `./config`:

### Демо

Не гарантирую что приложение будет жить вечно, но пользоваться им можешь [тут][demo].

## Changes log

Changes log can be [found here][link_changes_log].

## Support

If you will find any package errors, please, [make an issue][link_create_issue] in current repository.

## License

This is open-sourced software licensed under the [MIT License][link_license].

[badge_packagist_version]:https://img.shields.io/packagist/v/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&maxAge=180
[badge_php_version]:https://img.shields.io/packagist/php-v/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&longCache=true
[badge_ppm]:https://raw.githubusercontent.com/php-pm/ppm-badge/master/ppm-badge.png
[badge_license]:https://img.shields.io/packagist/l/tarampampam/mikrotik-hosts-parser.svg?style=flat-square&longCache=true
[link_packagist]:https://packagist.org/packages/tarampampam/mikrotik-hosts-parser
[link_ppm]:https://github.com/php-pm/php-pm
[link_changes_log]:https://github.com/tarampampam/mikrotik-hosts-parser/blob/php-version/CHANGELOG.md
[link_create_issue]:https://github.com/tarampampam/mikrotik-hosts-parser/issues/new/choose
[link_license]:https://github.com/tarampampam/mikrotik-hosts-parser/blob/php-version/LICENSE
[getcomposer]:https://getcomposer.org/download/
[demo]: https://stopad.hook.sh/
[habr]: https://habrahabr.ru/post/264001/
