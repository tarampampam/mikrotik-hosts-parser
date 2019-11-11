<template>
    <div>
        <pre ref="source-original" v-show="false">
## StopAD - Script for blocking advertisements, based on your defined hosts files<span v-if="serviceLink">
## For changing any parameters, please, use this link: {{ serviceLink }}</span>
##
## @github    &lt;{{ projectLink }}&gt;
## @version   {{ version }}
##
## Setup this Policy for script: [X] Read [X] Write [X] Policy [X] Test

:local hostScriptUrl "{{ scriptUri }}";
:local scriptName "{{ scriptName }}";
:local backupFileName "{{ backupFileName }}";
:local logPrefix "{{ logPrefix }}";

do {
  /tool fetch <span v-if="useSsl">check-certificate=no mode=https</span><span v-else>mode=http</span> url=$hostScriptUrl dst-path=("./".$scriptName);
  :delay 3s;
  :if ([:len [/file find name=$scriptName]] > 0) do={
    /system backup save name=$backupFileName;
    :delay 1s;
    :if ([:len [/file find name=($backupFileName.".backup")]] > 0) do={
      /ip dns static remove [/ip dns static find comment={{ entriesComment }}];
      /import file-name=$scriptName;
      /file remove $scriptName;
      :log info "$logPrefix AD block script imported, backup file (\"$backupFileName.backup\") created";
    } else={
      :log warning "$logPrefix Backup file not created, importing AD block script stopped";
    }
  } else={
    :log warning "$logPrefix AD block script not downloaded, script stopped";
  }
} on-error={
  :log warning "$logPrefix AD block script download FAILED";
};</pre>
        <pre class="mb-0"><code class="routeros" ref="source-render"></code></pre>
    </div>
</template>

<script>
    /* global module */
    /* global hljs */

    module.exports = {
        props: {
            serviceLink: {
                type: String
            },
            projectLink: {
                default: 'https://github.com/tarampampam/mikrotik-hosts-parser',
                type: String
            },
            version: {
                default: 'UNDEFINED',
                type: String
            },
            scriptUri: {
                default: 'UNDEFINED',
                type: String
            },
            scriptName: {
                default: 'stop_ad.script',
                type: String
            },
            backupFileName: {
                default: 'before_stopad',
                type: String
            },
            logPrefix: {
                default: '[StopAD]',
                type: String
            },
            entriesComment: {
                default: 'ADBlock',
                type: String
            },
            useSsl: {
                default: true,
                type: Boolean
            },
        },

        watch: {
            '$props': {
                handler: function () {
                    this.$nextTick(function () {
                        const src = this.$refs['source-original'], target = this.$refs['source-render'];

                        target.innerHTML = src.innerHTML; // make html code copying
                        hljs.highlightBlock(target)
                    });
                },
                immediate: true,
                deep: true,
            }
        },
    }
</script>

<style scoped>
    pre, pre code {
        background-color: transparent;
        word-wrap: break-word;
        white-space: pre-wrap;
    }
</style>
