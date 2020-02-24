<p align="center">
  <img src="https://hsto.org/webt/rx/1t/zd/rx1tzde8lrw8gqijqzdayj1gz1g.png" alt="Logo" width="128" />
</p>

# MikroTik hosts parser

![Release version][badge_release_version]
![Project language][badge_language]
[![Build Status][badge_build]][link_build]
[![Coverage][badge_coverage]][link_coverage]
[![Go Report][badge_goreport]][link_goreport]
[![Docker Build][badge_docker_build]][link_docker_hub]
[![License][badge_license]][link_license]

This application provides HTTP server, that can generate script for RouterOS-based routers for blocking "AD" hosts.

More information can be [found here][link_habr_post].

> Previous version (PHP) can be found in [`php-version` branch](https://github.com/tarampampam/mikrotik-hosts-parser/tree/php-version).

## Usage

For local application starting using binary file, you must compile application _(after repository cloning)_ using `GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./mikrotik-hosts-parser .` or `make build`, and then execute:

```bash
$ ./mikrotik-hosts-parser serve \
    --config ./serve.yml \
    --listen 0.0.0.0 \
    --port 8080 \
    --resources-dir ./public
```

This command will start HTTP server using configuration from `./serve.yml` on port `8080` and use directory `./public` for serving static files. Configuration file well-documented, so, feel free to change any settings on your choice!

> Configuration file allows you to use environment variables with default values!

After that you can navigate your browser to `http://127.0.0.1:8080/` and you will see something like that:

<p align="center">
  <img src="https://hsto.org/webt/k-/2f/ju/k-2fju1fgkbrsujcv15f-msgx2w.png" alt="screenshot" width="880" />
</p>

Special endpoint `/script/source?sources_urls=...` generates RouterOS-based script using passed http-get parameters _(watch examples on index page)_.

## Using docker

Just execute in your terminal:

```bash
$ docker run --rm -p 8080:8080/tcp tarampampam/mikrotik-hosts-parser:X.X.X
```

Where `X.X.X` is image tag _(application version)_. All available docker image tags can be [found here][link_docker_hub_tags].

## Demo

I can't guarantee that this links will available forever, but you can use this application on following links:

- <https://stopad.hook.sh/>
- <https://stopad.cgood.ru/>

## Testing

For application testing we use built-in golang testing feature and `docker-ce` + `docker-compose` as develop environment. So, just write into your terminal after repository cloning:

```shell
$ make test
```

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

[badge_build]:https://img.shields.io/github/workflow/status/tarampampam/mikrotik-hosts-parser/build?maxAge=30&logo=github
[badge_coverage]:https://img.shields.io/codecov/c/github/tarampampam/mikrotik-hosts-parser/master.svg?maxAge=30
[badge_goreport]:https://goreportcard.com/badge/github.com/tarampampam/mikrotik-hosts-parser
[badge_release_version]:https://img.shields.io/github/release/tarampampam/mikrotik-hosts-parser.svg?maxAge=30
[badge_docker_build]:https://img.shields.io/docker/cloud/build/tarampampam/mikrotik-hosts-parser?maxAge=30&label=docker
[badge_language]:https://img.shields.io/badge/language-go_1.13-blue.svg?longCache=true
[badge_license]:https://img.shields.io/github/license/tarampampam/mikrotik-hosts-parser.svg?longCache=true
[badge_release_date]:https://img.shields.io/github/release-date/tarampampam/mikrotik-hosts-parser.svg?maxAge=180
[badge_commits_since_release]:https://img.shields.io/github/commits-since/tarampampam/mikrotik-hosts-parser/latest.svg?maxAge=45
[badge_issues]:https://img.shields.io/github/issues/tarampampam/mikrotik-hosts-parser.svg?maxAge=45
[badge_pulls]:https://img.shields.io/github/issues-pr/tarampampam/mikrotik-hosts-parser.svg?maxAge=45
[link_goreport]:https://goreportcard.com/report/github.com/tarampampam/mikrotik-hosts-parser

[link_coverage]:https://codecov.io/gh/tarampampam/mikrotik-hosts-parser
[link_build]:https://github.com/tarampampam/mikrotik-hosts-parser/actions
[link_docker_hub]:https://hub.docker.com/r/tarampampam/mikrotik-hosts-parser/
[link_docker_hub_tags]:https://hub.docker.com/r/tarampampam/mikrotik-hosts-parser/tags
[link_license]:https://github.com/tarampampam/mikrotik-hosts-parser/blob/master/LICENSE
[link_releases]:https://github.com/tarampampam/mikrotik-hosts-parser/releases
[link_commits]:https://github.com/tarampampam/mikrotik-hosts-parser/commits
[link_changes_log]:https://github.com/tarampampam/mikrotik-hosts-parser/blob/master/CHANGELOG.md
[link_issues]:https://github.com/tarampampam/mikrotik-hosts-parser/issues
[link_create_issue]:https://github.com/tarampampam/mikrotik-hosts-parser/issues/new/choose
[link_pulls]:https://github.com/tarampampam/mikrotik-hosts-parser/pulls

[link_habr_post]:https://habr.com/ru/post/264001/
