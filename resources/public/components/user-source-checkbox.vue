<template>
    <div class="custom-control custom-checkbox pb-2">
        <input
            type="checkbox"
            class="custom-control-input"
            v-model="checked"
            @change="$emit('update:checked', checked)"
            :id="ID"
        />

        <label class="custom-control-label w-100" :for="ID">
            <input
                class="form-control form-control-sm bg-transparent border-primary text-light"
                :class="[validated === true ? 'is-valid' : '', validated === false ? 'is-invalid' : '']"
                type="url"
                placeholder="https://example.com/hosts.txt"
                v-model="sourceUri"
                @change="validateSourceUri();$emit('update:sourceUri', sourceUri)"
                @keyup="validateSourceUri();$emit('update:sourceUri', sourceUri)"
            />
        </label>
    </div>
</template>

<script>
    /* global module */

    module.exports = {
        props: {
            ID: {
                default: undefined,
                type: String
            },
            sourceUri: {
                required: true,
                type: String
            },
            checked: {
                default: false,
                type: Boolean
            },
        },

        data: /** @return {Object} */ function () {
            return {
                validated: undefined,
            }
        },

        mounted: function () {
            this.ID = this.ID || 'source_' + this._uid
        },

        methods: {
            validateSourceUri: /** @return {boolean} */ function () {
                if (typeof this.sourceUri === "string") {
                    this.validated = this.sourceUri.match(/^https?:\/\/[a-z0-9-.]+\.[a-z]{2,4}\/?([^\s<>#%",{}\\|\\\^\[\]`]+)?$/) !== null;
                    return;
                }

                this.validated = false;
            },
        },
    }
</script>

<style scoped>
</style>
