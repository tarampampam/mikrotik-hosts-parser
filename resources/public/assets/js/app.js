/* jsDoc nice ref: <https://fettblog.eu/typescript-jsdoc-superpowers/> */

'use strict';

/** @typedef {Vue} Vue */
/** @typedef {Axios} axios */
/** @typedef {Object} httpVueLoader */

// @link <https://github.com/vuejs/vue-devtools/issues/190#issuecomment-264203810>
Vue.config.devtools = true;

// @link <https://github.com/FranckFreiburger/http-vue-loader/#api>
httpVueLoader.httpRequest = function (url) {
    return axios.get(url)
        .then(function (res) {
            return res.data;
        })
        .catch(function (err) {
            return Promise.reject(err.status);
        });
};

Vue.use(httpVueLoader);

new Vue({
    el: '#app',
    components: {
        'line-delimiter': 'url:components/line-delimiter.vue',
        'main-header': 'url:components/main-header.vue',
        'about': 'url:components/about.vue',
        'script-source': 'url:components/script-source.vue',
        'faq': 'url:components/faq.vue',
        'main-footer': 'url:components/main-footer.vue',
    },
    data: {},
    methods: {},
    created: function () {
        //
    },
});
