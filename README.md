<p align="center">
  <img src="https://hsto.org/webt/rx/1t/zd/rx1tzde8lrw8gqijqzdayj1gz1g.png" alt="Logo" width="128" />
</p>

# MikroTik hosts parser

![Release version][badge_release_version]
[![Build Status][badge_build]][link_build]
[![Coverage][badge_coverage]][link_coverage]
[![Image size][badge_size_latest]][link_docker_hub]
[![License][badge_license]][link_license]

This application provides HTTP server, that can generate script for RouterOS-based routers for blocking "AD" hosts using 3rd party host-lists (available by HTTP):

```routeros
## Limit: 5000
## Cache lifetime: 30m0s
## Format: routeros
## Redirect to: 127.0.0.1
## Sources list:
##  - <https://adaway.org/hosts.txt>
##  - <https://www.malwaredomainlist.com/hostslist/hosts.txt>
## Excluded hosts:
##  - broadcasthost
##  - ip6-allhosts
##  - ip6-allnodes
##  - ip6-allrouters
## Cache HIT for <https://adaway.org/hosts.txt> (expires after 25m55s)
## Cache miss for <https://www.malwaredomainlist.com/hostslist/hosts.txt>

/ip dns static
add address=127.0.0.1 comment="ADBlock" disabled=no name="1-1ads.com"
add address=127.0.0.1 comment="ADBlock" disabled=no name="101com.com"
add address=127.0.0.1 comment="ADBlock" disabled=no name="101order.com"
add address=127.0.0.1 comment="ADBlock" disabled=no name="123freeavatars.com"

# ...
```

Hosts file format ([example](https://cdn.jsdelivr.net/gh/tarampampam/mikrotik-hosts-parser@master/.hosts/basic.txt)):

```
# Any comments
127.0.0.1   1-1ads.com
127.0.0.1   101com.com 101order.com
0.0.0.0     123freeavatars.com
```

All what you need is:

- Start current application HTTP server
- Make an HTTP request to the script generator endpoint `/script/source?sources_urls=...` with all required parameters (like records limit, hosts file URLs, exclusion list and others)
- Generated script source execute on your RouterOS-based hardware

More information can be [found here][link_habr_post].

> Previous version (PHP) can be found in [`php-version` branch](https://github.com/tarampampam/mikrotik-hosts-parser/tree/php-version).

## Installing

Download latest binary file for your os/arch from [releases page][link_releases] or use our [docker image][link_docker_hub] ([ghcr.io][link_ghcr]). Also you may need in configuration file [`./configs/config.yml`](configs/config.yml) and [`./web`](web) directory content for web UI access.

## Usage

This application supports next sub-commands:

Sub-command   | Description
------------- | -----------
`serve`       | Start HTTP server
`healthcheck` | Health checker for the HTTP server (use case - docker healthcheck) _(hidden in CLI help)_
`version`     | Display application version

And global flags:

Flag              | Description
----------------- | -----------
`--verbose`, `-v` | Verbose output
`--debug`         | Debug output
`--log-json`      | Logs in JSON format

### HTTP server starting

`serve` sub-command allows to use next flags:

Flag                    | Description                              | Default value              | Environment variable
----------------------- | ---------------------------------------- | -------------------------- | --------------------
`--listen`, `-l`        | IP address to listen on                  | `0.0.0.0` (all interfaces) | `LISTEN_ADDR`
`--port`, `-p`          | TCP port number                          | `8080`                     | `LISTEN_PORT`
`--resources-dir`, `-r` | Path to the directory with public assets | `./web`                    | `RESOURCES_DIR`
`--config`, `-c`        | Config file path                         | `./configs/config.yml`     | `CONFIG_PATH`
`--caching-engine`      | Caching engine (`memory` or `redis`)     | `memory`                   | `CACHING_ENGINE`
`--cache-ttl`           | Cached entries lifetime (examples: `50s`, `1h30m`) | `30m`            | `CACHE_TTL`
`--redis-dsn`           | Redis server DSN, required only if `redis` caching engine is enabled | `redis://127.0.0.1:6379/0` | `REDIS_DSN`

> Environment variables have higher priority then flag values.

Server starting command example:

```shell
$ ./mikrotik-hosts-parser serve \
    --config ./configs/config.yml \
    --listen 0.0.0.0 \
    --port 8080 \
    --resources-dir ./web
```

This command will start HTTP server using configuration from `./configs/config.yml` on TCP port `8080` and use directory `./web` for serving static files. Configuration file well-documented, so, feel free to change any settings on your choice!

> Configuration file allows you to use environment variables with default values ([used library](https://github.com/a8m/envsubst))!

After that you can navigate your browser to `http://127.0.0.1:8080/` and you will see something like that:

<p align="center">
  <img src="https://hsto.org/webt/k-/2f/ju/k-2fju1fgkbrsujcv15f-msgx2w.png" alt="screenshot" width="880" />
</p>

Special endpoint `/script/source?sources_urls=...` generates RouterOS-based script using passed http-get parameters _(watch examples on index page)_.

### Using docker

[![image stats](https://dockeri.co/image/tarampampam/mikrotik-hosts-parser)][link_docker_hub]

> All supported image tags [can be found here][link_docker_hub] and [here][link_ghcr].

Just execute in your terminal:

```shell
$ docker run --rm -p 8080:8080/tcp tarampampam/mikrotik-hosts-parser:X.X.X
```

Where `X.X.X` is image tag _(application version)_.

## Demo

I can't guarantee that this links will available forever, but you can use this application by the following links:

- <https://stopad.hook.sh/>
- <https://stopad.cgood.ru/>

## Testing

For application testing and building we use built-in golang testing feature and `docker-ce` + `docker-compose` as develop environment. So, just write into your terminal after repository cloning:

```shell
$ make test
```

Or build the binary file:

```shell
$ make build
```

## Releasing

New versions publishing is very simple - just make required changes in this repository, update [changelog file](CHANGELOG.md) and "publish" new release using repo releases page.

Binary files and docker images will be build and published automatically.

> New release will overwrite the `latest` docker image tag in both registers.

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

[badge_build]:https://img.shields.io/github/workflow/status/tarampampam/mikrotik-hosts-parser/tests?maxAge=30
[badge_coverage]:https://img.shields.io/codecov/c/github/tarampampam/mikrotik-hosts-parser/master.svg?maxAge=30
[badge_release_version]:https://img.shields.io/github/release/tarampampam/mikrotik-hosts-parser.svg?maxAge=30
[badge_size_latest]:https://img.shields.io/docker/image-size/tarampampam/mikrotik-hosts-parser/latest?maxAge=30
[badge_license]:https://img.shields.io/github/license/tarampampam/mikrotik-hosts-parser.svg?longCache=true
[badge_release_date]:https://img.shields.io/github/release-date/tarampampam/mikrotik-hosts-parser.svg?maxAge=180
[badge_commits_since_release]:https://img.shields.io/github/commits-since/tarampampam/mikrotik-hosts-parser/latest.svg?maxAge=45
[badge_issues]:https://img.shields.io/github/issues/tarampampam/mikrotik-hosts-parser.svg?maxAge=45
[badge_pulls]:https://img.shields.io/github/issues-pr/tarampampam/mikrotik-hosts-parser.svg?maxAge=45

[link_coverage]:https://codecov.io/gh/tarampampam/mikrotik-hosts-parser
[link_build]:https://github.com/tarampampam/mikrotik-hosts-parser/actions
[link_docker_hub]:https://hub.docker.com/r/tarampampam/mikrotik-hosts-parser/
[link_ghcr]:https://github.com/users/tarampampam/packages/container/package/mikrotik-hosts-parser
[link_docker_hub_tags]:https://hub.docker.com/r/tarampampam/mikrotik-hosts-parser/tags
[link_license]:https://github.com/tarampampam/mikrotik-hosts-parser/blob/master/LICENSE
[link_releases]:https://github.com/tarampampam/mikrotik-hosts-parser/releases
[link_commits]:https://github.com/tarampampam/mikrotik-hosts-parser/commits
[link_changes_log]:https://github.com/tarampampam/mikrotik-hosts-parser/blob/master/CHANGELOG.md
[link_issues]:https://github.com/tarampampam/mikrotik-hosts-parser/issues
[link_create_issue]:https://github.com/tarampampam/mikrotik-hosts-parser/issues/new/choose
[link_pulls]:https://github.com/tarampampam/mikrotik-hosts-parser/pulls

[link_habr_post]:https://habr.com/ru/post/264001/
