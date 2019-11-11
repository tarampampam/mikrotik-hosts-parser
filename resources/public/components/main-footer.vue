<template>
    <footer class="page-footer font-small text-center mb-4 mt-4">
        <a :href="link" target="_blank" v-html="text"/>
    </footer>
</template>

<script>
    /* global module */
    module.exports = {
        props: {
            link: {
                default: 'https://github.com/tarampampam/mikrotik-hosts-parser',
                type: String
            },
            text: {
                default: `Git <i class="fab fa-github"></i> Hub`,
                type: String
            },
        }
    }
</script>
