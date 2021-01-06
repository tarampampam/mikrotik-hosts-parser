<template>
    <div class="container">
        <div v-if="this.errored" class="alert alert-danger mt-3">
            <strong>Ooops!</strong> {{ this.errorMessage }}
        </div>

        <main-header :version="version"></main-header>

        <about></about>

        <div class="container" v-if="this.loaded">
            <hr class="delimiter"/>

            <fieldset class="form-group">
                <legend>
                    Источники
                    <button type="button"
                            class="btn btn-outline-info btn-sm border-primary ml-2"
                            v-on:click="addUserSource('', true)"
                            v-if="this.sources.length < this.maxSourcesCount">
                        <i class="fas fa-plus"></i> Добавить свой источник
                    </button>
                </legend>

                <div class="form-check pl-1">
                    <div
                        v-for="(source, index) in this.sources"
                        class="custom-control custom-checkbox pb-2"
                    >
                        <input
                            type="checkbox"
                            class="custom-control-input"
                            :id="index + '_source'"
                            v-model="source.isChecked"
                        />
                        <label
                            class="custom-control-label text-light"
                            :for="index + '_source'"
                            :class="{ 'w-100': source.isUserDefined }"
                        >
                            <div v-if="source.isUserDefined">
                                <input
                                    class="form-control form-control-sm bg-transparent border-primary text-light"
                                    type="url"
                                    placeholder="https://example.com/hosts.txt"
                                    v-model="source.uri"
                                    @change="validateSourceUri"
                                    @keyup="validateSourceUri"
                                />
                            </div>
                            <div v-else>
                                {{ source.name }}
                                <span v-if="source.count && source.count > 0" class="badge badge-primary">~{{ source.count }} записей</span>

                                <small id="fileHelp" class="form-text text-muted mt-0">
                                    <span v-if="source.description">{{ source.description }} &mdash; </span>
                                    <a :href="source.uri" target="_blank" class="text-muted">
                                        <code v-text="source.uri"></code> <i class="fas fa-external-link-alt small"></i>
                                    </a>
                                </small>
                            </div>
                        </label>
                    </div>
                </div>
            </fieldset>

            <div class="row">
                <div class="col-lg-6">
                    <fieldset class="form-group">
                        <legend class="h5">
                            Адрес перенаправления
                        </legend>
                        <div class="form-check pl-0">
                            <div class="form-group">
                                <input type="text"
                                       id="redirectIp"
                                       class="form-control form-control-sm bg-transparent border-primary text-light"
                                       v-model="redirectIp.value"
                                       placeholder="127.0.0.1"
                                       @change="validateRedirectIp"
                                       @keyup="validateRedirectIp"
                                />
                                <label class="form-text text-muted" for="redirectIp">
                                    Укажите IP (v4 или v6) адрес, куда перенаправлять запросы
                                </label>
                            </div>
                        </div>
                    </fieldset>

                    <fieldset class="form-group">
                        <legend class="h5">
                            Лимит записей
                        </legend>
                        <div class="form-check pl-0">
                            <div class="form-group">
                                <input type="number"
                                       min="0"
                                       max="100000000"
                                       id="recordsLimit"
                                       class="form-control form-control-sm bg-transparent border-primary text-light"
                                       v-model.number="recordsLimit"
                                       placeholder="0"
                                />
                                <label class="form-text text-muted" for="recordsLimit">
                                    Укажите максимальное количество возвращаемых записей
                                </label>
                            </div>
                        </div>
                    </fieldset>
                </div>

                <div class="col-lg-6">
                    <fieldset class="form-group">
                        <legend class="h5">
                            Исключения
                        </legend>
                        <div class="form-check pl-0">
                            <div class="form-group">
                                <textarea
                                    class="form-control bg-transparent border-primary text-light p-1 pl-2 pr-2 pb-2"
                                    id="excludesList"
                                    placeholder="adserver.yahoo.com"
                                    rows="6"
                                    @change="updateExcludesList"
                                    @keyup="updateExcludesList"
                                >{{ excludesList.join('\n') }}</textarea>
                                <label class="form-text text-muted" for="excludesList">
                                    Можете указать те хосты, которые необходимо исключить из итогового скрипта,
                                    одна строка для одного хоста
                                </label>
                            </div>
                        </div>
                    </fieldset>
                </div>
            </div>

            <fieldset class="form-group">
                <legend>
                    Адрес скрипта
                </legend>
                <div class="form-check pl-3 pr-3">
                    <a :href="getScriptGeneratorUri()" target="_blank">
                        <code
                            class="font-weight-bolder" style="word-break: break-all"
                        >{{ getScriptGeneratorUri() }}</code> <i class="fas fa-external-link-alt small"></i>
                    </a>
                </div>
            </fieldset>

            <fieldset class="form-group">
                <legend class="h3">
                    Скрипт для маршрутизатора
                </legend>
                <div class="form-check pl-1">
                    <script-source
                        :service-link="window.location.toString()"
                        :version="version"
                        :script-uri="getScriptGeneratorUri()"
                        :use-ssl="useSsl"
                        :entries-comment="entriesComment"
                    ></script-source>
                </div>
            </fieldset>

            <hr class="delimiter"/>
        </div>
        <div v-else>
            <div class="w-100 text-center mt-5 mb-5">
                <span class="spinner-border spinner-border-sm mr-1"></span> Загрузка..
            </div>
        </div>

        <faq
            :cache-lifetime-sec="cacheLifetimeSec"
            :max-sources-count="maxSourcesCount"
            :max-source-size-kb="maxSourceSizeBytes / 1024"
        ></faq>

        <main-footer></main-footer>
    </div>
</template>

<script>
    'use strict';

    /* global module */
    /* global axios */
    /* global hljs */

    module.exports = {
        components: {
            'main-header': 'url:components/main-header.vue',
            'about': 'url:components/about.vue',
            'script-source': 'url:components/script-source.vue',
            'faq': 'url:components/faq.vue',
            'main-footer': 'url:components/main-footer.vue',
        },

        data: function () {
            return {
                loaded: false,
                errored: false,
                errorMessage: 'Something went wrong',
                maxSourcesCount: 25,
                maxSourceSizeBytes: 1024,
                sources: [],
                redirectIp: {
                    value: '0.0.0.0',
                    is_valid: true
                },
                recordsLimit: 5000,
                excludesList: [
                    'localhost',
                    'localhost.localdomain',
                    'broadcasthost',
                    'local',
                ],
                version: 'UNKNOWN_VERSION',
                format: 'routeros',
                cacheLifetimeSec: NaN,
                entriesComment: 'ADBlock',
                scriptGeneratorPath: 'script/source',
                useSsl: window.location.protocol === 'https:',
            }
        },

        methods: {
            newSource:
                /**
                 * Source object factory.
                 *
                 * @param {string} uri
                 * @param {string} name
                 * @param {number} count
                 * @param {string} desc
                 * @param {boolean} isChecked
                 * @param {boolean} isUserDefined
                 *
                 * @throws {Error} If required parameters was not passed.
                 *
                 * @returns {Source}
                 */
                function (uri, name, count, desc, isChecked, isUserDefined) {
                    if (typeof uri !== 'string') {
                        throw Error('Required arguments for factory was not passed');
                    }

                    /**
                     * @typedef {Object} Source
                     * @property {string}  uri Source URI
                     * @property {string}  name Human-like source name
                     * @property {number}  count Approximate source entries count
                     * @property {string}  description Human-like source description
                     * @property {boolean} isChecked Checked state
                     * @property {boolean} isUserDefined Is source defined by user?
                     */
                    return {
                        uri: uri.trim(),
                        name: typeof name === "string" ? name.trim() : undefined,
                        count: typeof count === "number" ? count : NaN,
                        description: typeof desc === "string" ? desc.trim() : undefined,
                        isChecked: typeof isChecked === "boolean" ? isChecked : false,
                        isUserDefined: typeof isUserDefined === "boolean" ? isUserDefined : false,
                    };
                },

            addUserSource:
                /**
                 * @param {string} sourceUri
                 * @param {boolean} isChecked
                 */
                function (sourceUri, isChecked) {
                    this.sources.push(this.newSource(
                        sourceUri, undefined, undefined, undefined, isChecked, true
                    ));
                },

            updateExcludesList:
                /**
                 * @param {Event} event
                 */
                function (event) {
                    const res = [];

                    if (typeof event.target.value === "string") {
                        event.target.value.split("\n").forEach(/** @param {string} line */ function (line) {
                            line = line
                                .trim()
                                .replace(/\s\s+/g, ' ')
                                .replace(/[^a-zа-яё0-9\*-_\.\s:]/gi, '');
                            if (line.length > 0 && !line.includes(' ')) {
                                res.push(line);
                            }
                        })
                    }

                    this.excludesList = res;
                },

            validateSourceUri:
                /**
                 * @param {Event} event
                 */
                function (event) {
                    let validated = false;
                    const $el = event.target,
                        validClass = 'is-valid',
                        invalidClass = 'is-invalid';

                    if (typeof $el.value === "string") {
                        validated = $el.value.match(/^https?:\/\/[a-z0-9-.]+\.[a-z]{2,4}\/?([^\s<>#%",{}\\|\\\^\[\]`]+)?$/) !== null;
                    }

                    if (validated === true) {
                        $el.classList.add(validClass);
                        $el.classList.remove(invalidClass);
                    } else {
                        $el.classList.add(invalidClass);
                        $el.classList.remove(validClass);
                    }
                },

            validateRedirectIp:
                /**
                 * @param {Event} event
                 */
                function (event) {
                    const $el = event.target,
                        validClass = 'is-valid',
                        invalidClass = 'is-invalid';

                    this.redirectIp.value = $el.value
                        .trim()
                        .replace(/\s\s+/g, ' ')
                        .replace(/[^0-9a-f:\.]/ig, '');

                    this.redirectIp.is_valid = typeof this.redirectIp.value === "string" && this.validateIp(this.redirectIp.value);

                    if (this.redirectIp.is_valid === true) {
                        $el.classList.add(validClass);
                        $el.classList.remove(invalidClass);
                    } else {
                        $el.classList.add(invalidClass);
                        $el.classList.remove(validClass);
                    }
                },

            getScriptGeneratorUri:
                /**
                 * Get script generator URI.
                 *
                 * @return {string}
                 */
                function () {
                    let location = window.location.toString();
                    let baseUri = location.substring(0, location.lastIndexOf('/'))
                        + '/' + this.scriptGeneratorPath.toString().replace(/^\//, '');
                    let params = this.getScriptUriParams();

                    return baseUri + (params.length > 0 ? '?' + params : '');
                },

            getScriptUriParams:
                /**
                 * Build script generation URI params.
                 *
                 * @return {string}
                 */
                function () {
                    let parts = {
                            format: this.format,
                            version: this.version,
                        },
                        recordsLimit = parseInt(this.recordsLimit, 10);

                    parts['redirect_to'] = this.redirectIp.is_valid ? this.redirectIp.value : null;
                    parts['limit'] = recordsLimit > 0 ? recordsLimit : null;
                    parts['sources_urls'] = this.sources
                        .map(/** @param {Source} source */ function (source) {
                            if (source.isChecked === true && source.uri !== '') {
                                return encodeURIComponent(source.uri);
                            }

                            return null;
                        })
                        .filter(/** @param {?Source} source */ function (source) {
                            return source != null;
                        })
                        .join(',');
                    parts['excluded_hosts'] = this.excludesList
                        .map(/** @param {string} value */ function (value) {
                            return encodeURIComponent(value.toString().trim());
                        })
                        .join(',');

                    let partsArray = [];

                    // Make clean
                    for (let part_name in parts) {
                        if (
                            parts[part_name] !== null && parts[part_name] !== undefined && parts[part_name] !== []
                            || (typeof parts[part_name] === 'string' && parts[part_name].length > 0)
                        ) {
                            partsArray.push(part_name + '=' + parts[part_name]);
                        }
                    }

                    return partsArray.join('&');
                },

            validateIp:
                /**
                 * Validates passed string against IP address regExp.
                 *
                 * @param {string} value
                 * @return {boolean}
                 */
                function (value) {
                    // @link https://www.regextester.com/104038
                    const regExp = /((^\s*((([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]))\s*$)|(^\s*((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?\s*$))/;

                    return regExp.test(value);
                },
        },

        mounted: function () {
            const self = this;

            /**
             * @typedef {Object} AxiosResponse
             * @property {Object} data       Response data
             * @property {number} status     HTTP status code
             * @property {string} statusText HTTP status message
             * @property {Object} headers    Headers that the server responded with. All header names are lower cased
             * @property {Object} config     Config that was provided to `axios` for the request
             * @property {Object} request    Request that generated this response
             */

            axios
                .all([
                    axios.request({method: 'get', url: '/api/version', timeout: 2000}),
                    axios.request({method: 'get', url: '/api/routes', timeout: 2000}),
                    axios.request({method: 'get', url: '/api/settings', timeout: 2000}),
                ])
                .then(axios.spread(
                    /**
                     * @param {AxiosResponse} versionResponse
                     * @param {AxiosResponse} routesResponse
                     * @param {AxiosResponse} settingsResponse
                     */
                    function (versionResponse, routesResponse, settingsResponse) {
                        // Attach `hasOwnNestedProperty` function into each `*Response.data` object
                        [versionResponse, routesResponse, settingsResponse].forEach(function (response) {
                            // @link: <https://stackoverflow.com/a/33445095>
                            response.data.hasOwnNestedProperty = /** @param {string} path */ function (path) {
                                if (typeof path !== "string" || path.length <= 0) {
                                    return false;
                                }

                                for (let i = 0, properties = path.split('.'), obj = this; i < properties.length; i++) {
                                    let prop = properties[i];

                                    if (!obj || !obj.hasOwnProperty(prop)) {
                                        return false;
                                    } else {
                                        obj = obj[prop];
                                    }
                                }

                                return true;
                            };
                        });

                        // Append sources
                        if (settingsResponse.data.hasOwnNestedProperty('sources.provided')) {
                            /**
                             * @typedef {Object} RawSourceData
                             * @property {string} uri
                             * @property {string} name
                             * @property {string} description
                             * @property {boolean} default
                             * @property {number} count
                             */
                            settingsResponse.data.sources.provided.forEach(/** @param {RawSourceData} s */ function (s) {
                                self.sources.push(self.newSource(
                                    s.uri,
                                    s.name,
                                    s.count,
                                    s.description,
                                    s.default,
                                ));
                            });
                        }

                        if (settingsResponse.data.hasOwnNestedProperty('sources.max')) {
                            self.maxSourcesCount = parseInt(settingsResponse.data.sources.max, 10);
                        }

                        if (settingsResponse.data.hasOwnNestedProperty('sources.max_source_size')) {
                            self.maxSourceSizeBytes = parseInt(settingsResponse.data.sources.max_source_size, 10);
                        }

                        if (settingsResponse.data.hasOwnNestedProperty('redirect.addr')) {
                            self.redirectIp.value = settingsResponse.data.redirect.addr;
                        }

                        if (settingsResponse.data.hasOwnNestedProperty('excludes.hosts')) {
                            self.excludesList = settingsResponse.data.excludes.hosts;
                        }

                        if (settingsResponse.data.hasOwnNestedProperty('records.comment')) {
                            self.entriesComment = settingsResponse.data.records.comment;
                        }

                        if (settingsResponse.data.hasOwnNestedProperty('cache.lifetime_sec')) {
                            self.cacheLifetimeSec = settingsResponse.data.cache.lifetime_sec;
                        }

                        if (versionResponse.data.hasOwnNestedProperty('version')) {
                            self.version = versionResponse.data.version;
                        }

                        if (routesResponse.data.hasOwnNestedProperty('script_generator.path')) {
                            self.scriptGeneratorPath = routesResponse.data.script_generator.path;
                        }
                    }
                ))
                .catch(/** @param {Error} error */ function (error) {
                    self.errored = true;
                    self.errorMessage = error.message;
                })
                .finally(function () {
                    self.$nextTick(function () {
                        // Code that will run only after the entire view has been rendered
                        self.loaded = true;
                    });
                });
        },
    }
</script>

<style scoped>
    hr.delimiter {
        border: none;
        height: 2px;
        background-image: linear-gradient(to right, #272B30, #2d3238, #272B30);
        margin: 2em 0 1.5em;
    }

    pre {
        background-color: transparent;
    }
</style>
