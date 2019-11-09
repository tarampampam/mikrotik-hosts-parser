<template>
    <div class="custom-control custom-checkbox pb-2">
        <input
            type="checkbox"
            class="custom-control-input"
            :id="ID"
            v-model="checked"
            @change="$emit('update:checked', checked)"
        />
        <label class="custom-control-label text-light" :for="ID">
            {{ sourceName }}
            <span v-if="entriesCount" class="badge badge-primary">~{{ entriesCount }} записей</span>
        </label>

        <small id="fileHelp" class="form-text text-muted mt-0">
            <span v-if="description">{{ description }} &mdash; </span>
            <a
                :href="sourceUri"
                target="_blank"
                class="text-muted"
            >
                <code>{{ sourceUri }}</code> <i class="fas fa-external-link-alt small"></i>
            </a>
        </small>
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
            sourceName: {
                required: true,
                type: String
            },
            description: {
                type: String
            },
            sourceUri: {
                required: true,
                type: String
            },
            entriesCount: {
                default: NaN,
                type: Number
            },
            checked: {
                default: false,
                type: Boolean
            },
        },
        mounted: function () {
            this.ID = this.ID || 'source_' + this._uid
        }
    }
</script>

<style scoped>
</style>
