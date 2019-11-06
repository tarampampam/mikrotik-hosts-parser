<template>
    <div class="container">
        <div v-if="errored" class="alert alert-danger mt-3">
            <strong>Ooops!</strong> {{ errorMessage }}.
        </div>

        <div v-if="!loaded" class="w-100 text-center mt-5">
            <div class="spinner-border text-secondary" role="status">
                <span class="sr-only">Loading...</span>
            </div>
        </div>

        <div v-else>
            <main-header></main-header>

            <about></about>

            <line-delimiter></line-delimiter>

            <script-source
                service-link=""
                version=""
                script-uri=""
                :use-ssl="false"
            ></script-source>

            <line-delimiter></line-delimiter>

            <faq></faq>

            <main-footer></main-footer>
        </div>
    </div>
</template>

<script>
    /* global module */
    module.exports = {
        components: {
            'line-delimiter': 'url:components/line-delimiter.vue',
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
            }
        },
        methods: {},
        created: function () {
            //
        },
        mounted: function () {
            this.$nextTick(function () {
                // Code that will run only after the entire view has been rendered
                this.loaded = true;
            })
        },
    }
</script>

<style>
</style>
