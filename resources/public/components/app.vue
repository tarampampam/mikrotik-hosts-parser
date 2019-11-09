<template>
    <div class="container">
        <div v-if="errored" class="alert alert-danger mt-3">
            <strong>Ooops!</strong> {{ errorMessage }}.
        </div>

        <main-header></main-header>

        <about></about>

        <line-delimiter></line-delimiter>

        <div class="container" v-if="loaded">
            <fieldset class="form-group">
                <legend>
                    Источники
                    <button type="button"
                            class="btn btn-outline-info btn-sm border-primary ml-2"
                            v-on:click="addUserSource('', true)"
                            v-if="sources.length <= maxSourcesCount">
                        <i class="fas fa-plus"></i> Добавить свой источник
                    </button>
                </legend>

                <div class="form-check pl-1">
                    <source-checkbox
                        v-for="source in sources"
                        v-if="!source.isUserDefined"
                        :source-name="source.name"
                        :entries-count="source.count"
                        :description="source.description"
                        :source-uri="source.uri"
                        :checked.sync="source.isChecked"
                    ></source-checkbox>

                    <user-source-checkbox
                        v-for="source in sources"
                        v-if="source.isUserDefined"
                        :source-uri.sync="source.uri"
                        :checked.sync="source.isChecked"
                    ></user-source-checkbox>
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
                                       v-model="redirectIp"
                                       placeholder="127.0.0.1"
                                />
                                <label class="form-text text-muted"
                                       for="redirectIp">Укажите IP (v4) адрес, куда перенаправлять запросы</label>
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
                                       id="recordsLimit"
                                       class="form-control form-control-sm bg-transparent border-primary text-light"
                                       v-model="recordsLimit"
                                       placeholder="0"
                                />
                                <label class="form-text text-muted"
                                       for="recordsLimit">Укажите максимальное количество возвращаемых записей</label>
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
                                    class="form-control bg-transparent border-primary text-light p-1 pl-2 pr-2 pb-2 min"
                                    id="excludesList"
                                    placeholder="adserver.yahoo.com"
                                    rows="6"
                                    @change="updateExcludesList"
                                    @keyup="updateExcludesList"
                                >{{ excludesList.join('\n') }}</textarea>
                                <label class="form-text text-muted"
                                       for="excludesList">
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
                    <h3>Скрипт для маршрутизатора</h3>
                </legend>
                <div class="form-check pl-1">
                    <script-source
                        service-link=""
                        version=""
                        script-uri=""
                        :use-ssl="false"
                    ></script-source>
                </div>
            </fieldset>

            <line-delimiter></line-delimiter>
        </div>

        <faq></faq>

        <main-footer></main-footer>
    </div>
</template>

<script>
    /* global module */
    /* global axios */

    const clean = new function () {
        /**
         * Make string cleaning.
         *
         * @param {string} string
         * @returns {string}
         */
        this.string = function (string) {
            return string
                .trim()
                .replace(/\s\s+/g, ' ')
                .replace(/[^a-zа-яё0-9\*-_\.\s:]/gi, '');
        };

        /**
         * Make IP address clean.
         *
         * @param {string} string
         * @returns {string}
         */
        this.ip = function (string) {
            return string
                .trim()
                .replace(/\s\s+/g, ' ')
                .replace(/[^0-9a-z:\.]/g, '');
        };
    };

    /**
     * Source object.
     *
     * @typedef {Object} Source
     * @property {string}  name Human-like source name
     * @property {string}  uri Source URI
     * @property {number}  count Approximate source entries count
     * @property {string}  description Human-like source description
     * @property {boolean} isChecked Checked state
     * @property {boolean} isUserDefined Is source defined by user?
     */

    /**
     * Source object factory.
     *
     * @param {string} name
     * @param {string} uri
     * @param {number} count
     * @param {string} desc
     * @param {boolean} isChecked
     * @param {boolean} isUserDefined
     *
     * @throws {Error} If required parameters not passed.
     *
     * @returns {Source}
     */
    let sourceFactory = function (name, uri, count, desc, isChecked, isUserDefined) {
        if (typeof uri === 'undefined') {
            throw Error('Required arguments for factory was not passed');
        }

        return {
            name: name,
            uri: uri,
            count: count,
            description: desc,
            isChecked: isChecked || false,
            isUserDefined: isUserDefined || false,
        };
    };

    module.exports = {
        components: {
            'line-delimiter': 'url:components/line-delimiter.vue',
            'main-header': 'url:components/main-header.vue',
            'about': 'url:components/about.vue',
            'source-checkbox': 'url:components/source-checkbox.vue',
            'user-source-checkbox': 'url:components/user-source-checkbox.vue',
            'script-source': 'url:components/script-source.vue',
            'faq': 'url:components/faq.vue',
            'main-footer': 'url:components/main-footer.vue',
        },

        /**
         * @typedef {Object} AppData
         * @property {boolean} loaded Loading completed?
         * @property {boolean} errored Some fatal error occurred?
         * @property {string} errorMessage Fatal error message
         * @property {number} maxSourcesCount Maximum sources count
         * @property {Source[]} sources Source definition objects
         * @property {string} redirectIp
         * @property {number} recordsLimit
         */
        data: /** @return {AppData} */ function () {
            return {
                loaded: false,
                errored: false,
                errorMessage: 'Something went wrong',
                maxSourcesCount: 25,
                sources: [],
                redirectIp: '0.0.0.0',
                recordsLimit: 5000,
                excludesList: [
                    'localhost',
                    'localhost.localdomain',
                    'broadcasthost',
                    'local',
                    'ip6-localhost',
                    'ip6-loopback',
                    'ip6-localnet',
                    'ip6-mcastprefix',
                    'ip6-allnodes',
                    'ip6-allrouters',
                    'ip6-allhosts',
                ],
            }
        },

        methods: {
            addUserSource:
                /**
                 * @param {string} sourceUri
                 * @param {boolean} isChecked
                 */
                function (sourceUri, isChecked) {
                    this.sources.push(sourceFactory(
                        undefined, sourceUri, NaN, undefined, isChecked, true
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
                            line = clean.string(line);
                            if (line.length > 0) {
                                res.push(line);
                            }
                        })
                    }

                    this.excludesList = res;
                },
        },

        mounted: function () {
            const self = this;

            axios
                .request({method: 'get', url: 'https://httpbin.org/json', timeout: 5000})
                .then(function (response) {
                    self.sources.push(sourceFactory('Foo name', 'https://ya.ru/robots.txt', 123, 'Foo desc', true));
                    self.sources.push(sourceFactory('Bar name', 'https://ya.ru/robots.txt', 123, 'Foo desc', false));
                    self.sources.push(sourceFactory('Baz name', 'https://ya.ru/robots.txt', 123, 'Foo desc', true));
                    self.loaded = true;
                })
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
</style>
