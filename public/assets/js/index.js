"use strict";

(function ($) {

    /**
     * Script & URI to the script generator object.
     */
    var dataForUserGenerator = new function () {
        /**
         * Containers storage.
         *
         * @type {*}
         */
        this.$objects = {
            // Script URI container object
            script_uri: $('.script-uri'),
            // Script source container object
            script_source: $('.script-source'),
            // Sources checkboxes
            sources_checkboxes: $('.source-checkbox'),
            // User defined sources
            user_defined_sources: $('.user-defined-sources'),
            // "Redirect to" IP address
            redirect_to_ip: $('.redirect-to-ip').last(),
            // Result entries limit
            result_limit: $('.result-limit').last(),
            // Excluded hosts names
            excluded_hosts: $('.excluded-hosts')
        };

        /**
         * Router script source.
         *
         * @returns {string}
         */
        this.getScriptSourceCode = function () {
            return "## StopAD - Script for blocking advertisements, based on your defined hosts files\n\
## For changing any parameters, please, use this link: {%app_uri%}\n\
##\n\
## @github    &lt;{%repository_uri%}&gt;\n\
## @version   {%app_version%}\n\
##\n\
## Setup this Policy for script: [X] Read [X] Write [X] Policy [X] Test\n\
\n\
:local hostScriptUrl \"{%script_uri%}\";\n\
:local scriptName \"stop_ad.script\";\n\
:local backupFileName \"before_stopad\";\n\
:local logPrefix \"[StopAD]\";\n\
\n\
do {\n\
  /tool fetch {%fetch_mode%} url=$hostScriptUrl dst-path=(\"./\".$scriptName);\n\
  :if ([:len [/file find name=$scriptName]] > 0) do={\n\
    /system backup save name=$backupFileName;\n\
    :delay 1s;\n\
    :if ([:len [/file find name=($backupFileName.\".backup\")]] > 0) do={\n\
      /ip dns static remove [/ip dns static find comment={%ad_entries_comment%}];\n\
      /import file-name=$scriptName;\n\
      /file remove $scriptName;\n\
      :log info \"$logPrefix AD block script imported, backup file (\\\"$backupFileName.backup\\\") created\";\n\
    } else={\n\
      :log warning \"$logPrefix Backup file not created, importing AD block script stopped\";\n\
    }\n\
  } else={\n\
    :log warning \"$logPrefix AD block script not downloaded, script stopped\";\n\
  }\n\
} on-error={\n\
  :log warning \"$logPrefix AD block script download FAILED\";\n\
};";
        };

        /**
         * Enable code highlighting.
         */
        this.updateHighlightJs = function () {
            /**
             * Make highlight JS initialization.
             *
             * @param {object} block
             */
            var initializeHighlight = function (block) {
                hljs.highlightBlock(block);
            };
            this.$objects.script_uri.each(function (j, block) {
                initializeHighlight(block);
            });
            this.$objects.script_source.each(function (j, block) {
                initializeHighlight(block);
            });
        };

        /**
         * Get variable, passed to the page through javascript.
         *
         * @param {string} var_name
         * @returns {*}
         */
        this.getBackendVariable = function (var_name) {
            if (typeof var_name === 'string') {
                if (window.hasOwnProperty('backend_vars')) {
                    if (window.backend_vars.hasOwnProperty(var_name)) {
                        return window.backend_vars[var_name];
                    } else {
                        console.error('Backend variable "' + var_name + '" was not found');
                    }
                } else {
                    console.error('Object for storage backend variables was not found');
                }
            } else {
                throw new TypeError('Variable name must be string');
            }
        };

        /**
         * Make multiple replaces in the string.
         *
         * @param {string} string
         * @param {Array} search_for
         * @param {Array} replace_with
         * @returns {string}
         */
        this.makeReplacesInString = function (string, search_for, replace_with) {
            var open_tag = '{%',
                close_tag = '%}';
            if (typeof string === 'string' && $.isArray(search_for) && $.isArray(replace_with)) {
                if (search_for.length === replace_with.length) {
                    for (var i = 0, len = search_for.length; i < len; ++i) {
                        string = string.split(open_tag + search_for[i] + close_tag).join(replace_with[i]);
                    }
                    return string;
                } else {
                    throw new RangeError('"search_for" and "replace_with" lengths must be equals');
                }
            } else {
                throw new TypeError('Passed invalid data types');
            }
        };

        /**
         * Make URI validate.
         *
         * @param {string} url
         * @returns {boolean}
         */
        this.isValidUri = function (url) {
            return Boolean(url.match(/^(ht|f)tps?:\/\/[a-z0-9-.]+\.[a-z]{2,4}\/?([^\s<>#%",{}\\|\\\^\[\]`]+)?$/));
        };

        /**
         * Make string clean.
         *
         * @param {string} string
         * @returns {string}
         */
        this.makeStringClean = function (string) {
            return $.trim(string.replace(/\s\s+/g, ' ')).replace(/[^a-zа-яё0-9\*-_\.\s]/gi, '');
        };

        /**
         * Make URI string clean.
         *
         * @param {string} string
         * @returns {string}
         */
        this.makeCleanUri = function (string) {
            return $.trim(string.replace(/\s\s+/g, ' ')).replace(/[|;$%@"<>()+,]/g, '');
        };

        /**
         * Make IP address clean.
         *
         * @param {string} string
         * @returns {string}
         */
        this.makeCleanIpAddress = function (string) {
            return $.trim(string.replace(/\s\s+/g, ' ')).replace(/[^0-9a-z:\.]/g, '');
        };

        /**
         * Get the sources URIs array.
         *
         * @returns {Array}
         */
        this.getSourcesArray = function () {
            var self = this,
                result = [];

            // Read 'sources_checkboxes' values
            this.$objects.sources_checkboxes.filter(':checked').each(function (i, $input) {
                var uri = $($input).attr('data-url');
                if ((typeof uri !== 'undefined') && (uri !== false) && self.isValidUri(uri)) {
                    result.push(uri);
                }
            });

            // Read user-defines sources list
            this.$objects.user_defined_sources.each(function (i, $input) {
                $.each($($input).val().split("\n"), function (i, line) {
                    line = self.makeCleanUri(line);
                    if (self.isValidUri(line)) {
                        result.push(line);
                    }
                });
            });

            // Remove duplicates
            return result.filter(function (item, pos) {
                return result.indexOf(item) === pos;
            });
        };

        /**
         * Get excluded hosts array.
         *
         * @returns {Array}
         */
        this.getExcludedHostsArray = function () {
            var self = this, result = [];

            // Read user-defines excluded hosts
            this.$objects.excluded_hosts.each(function (i, $input) {
                $.each($($input).val().split("\n"), function (i, line) {
                    line = self.makeStringClean(line);
                    if (typeof line === 'string' && line.length > 0) {
                        result.push(line);
                    }
                });
            });

            // Remove duplicates
            return result.filter(function (item, pos) {
                return result.indexOf(item) === pos;
            });
        };

        /**
         * Build script URI parameters (as object).
         *
         * @returns {string}
         */
        this.buildScriptUriParameters = function () {
            var parts = {
                    format: '{%format%}',
                    version: '{%app_version%}'
                },
                parts_array = [];

            var redirect_to_ip = typeof this.$objects.redirect_to_ip === 'object'
                    ? this.makeCleanIpAddress(this.$objects.redirect_to_ip.val())
                    : '',
                result_limit = typeof this.$objects.result_limit === 'object'
                    ? parseInt(this.makeStringClean(this.$objects.result_limit.val()), 10)
                    : 0;

            parts['redirect_to'] = (typeof redirect_to_ip === 'string' && redirect_to_ip.length >= 5)
                ? redirect_to_ip
                : null;
            parts['limit'] = (typeof result_limit === 'number' && result_limit > 0)
                ? result_limit
                : null;
            parts['sources_urls'] = this.getSourcesArray().map(function (value) {
                return encodeURIComponent(value.trim());
            }).join(',');
            parts['excluded_hosts'] = this.getExcludedHostsArray().map(function (value) {
                return encodeURIComponent(value.trim());
            }).join(',');

            // Make clean
            for (var part_name in parts) {
                if (
                    parts[part_name] !== null && parts[part_name] !== undefined && parts[part_name] !== []
                    || (typeof parts[part_name] === 'string' && parts[part_name].length > 0)
                ) {
                    parts_array.push(part_name + '=' + parts[part_name]);
                }
            }

            return parts_array.join('&');
        };

        /**
         * Main method - update URI and script source.
         */
        this.updateData = function () {
            var script_uri,
                script_base_uri = this.getBackendVariable('SCRIPT_SOURCE_BASE_URI'),
                app_version = this.getBackendVariable('APP_VERSION') || 'unknown',
                script_source = this.getScriptSourceCode(),
                app_uri = window.location.protocol + '//' + window.location.hostname + window.location.pathname,
                protocol = (window.location.protocol === 'https:') ? 'https' : 'http';


            // Prepare script link URI
            script_uri = this.makeReplacesInString(script_base_uri + '?' + this.buildScriptUriParameters(), [
                'format',
                'app_version'
            ], [
                'routeros', // Hardcode, yes, i know
                app_version
            ]);

            // Make patterns replaces
            script_source = this.makeReplacesInString(script_source, [
                'app_uri',
                'app_version',
                'repository_uri',
                'script_uri',
                'ad_entries_comment',
                'fetch_mode'
            ], [
                app_uri,
                app_version,
                this.getBackendVariable('REPOSITORY_URI') || 'unknown',
                script_uri,
                this.getBackendVariable('SCRIPT_AD_ENTRIES_COMMENT') || 'ADBlock',
                protocol === 'https' ? 'check-certificate=no mode=https' : 'mode=http'
            ]);

            this.$objects.script_uri.html('<a href="' + script_uri + '" target="_blank">' + script_uri + '</a>');
            this.$objects.script_source.html(script_source);

            this.updateHighlightJs();
        };
    };

    /**
     * On UI elements change - make update.
     */
    $('input, textarea, button').on('change click keypress keyup', function () {
        dataForUserGenerator.updateData();
    });

    // Make direct event call
    $('input').first().keyup();

})(jQuery);