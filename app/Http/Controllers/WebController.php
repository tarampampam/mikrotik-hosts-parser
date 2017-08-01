<?php

namespace App\Http\Controllers;

/**
 * Class WebController
 */
class WebController extends AbstractController
{
    /**
     * @return \Illuminate\View\View
     */
    public function index()
    {
        return view('index', [
            'sources'                => config('sources'),
            'user_sources_limit'     => config('limits.user_sources'),
            'default_redirect_ip'    => config('defaults.redirect_ip'),
            'result_entries_limit'   => config('limits.result_entries'),
            'excluded_hosts'         => implode("\n", config('defaults.excluded_hosts', [])),
            'source_cache_lifetime'  => config('limits.source.cache.lifetime'),
            'source_uri_length'      => config('limits.source_uri_length'),
            'excluded_hosts_limit'   => config('limits.excluded_hosts'),
            'source_file_size_limit' => config('limits.source_file_size'),
            'sources_protocols'      => config('limits.sources_protocols'),

            'js_vars' => [
                'APP_VERSION'               => config('app.version'),
                'REPOSITORY_URI'            => config('contacts.repository.uri'),
                'SCRIPT_SOURCE_BASE_URI'    => route('script.source'),
                'SCRIPT_AD_ENTRIES_COMMENT' => config('defaults.script_ad_entries_comment'),
            ],
        ]);
    }
}
