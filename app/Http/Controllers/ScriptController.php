<?php

namespace App\Http\Controllers;

use Carbon\Carbon;
use Illuminate\Http\Request;
use App\Services\HostsParser\HostsParser;

/**
 * Class ScriptController.
 */
class ScriptController extends AbstractController
{
    /**
     * @var \Illuminate\Cache\Repository
     */
    protected $cache;

    /**
     * Response cache lifetime (in seconds);.
     *
     * @var int
     */
    protected $cache_lifetime = 240;

    /**
     * ScriptController constructor.
     */
    public function __construct()
    {
        $this->cache = app('cache');
    }

    /**
     * Generate script source.
     *
     * @param Request $request
     *
     * @return \Illuminate\Http\Response|\Laravel\Lumen\Http\ResponseFactory
     */
    public function source(Request $request)
    {
        $cache_key = $this->getHash($request->all());

        if ($this->cache->has($cache_key)) {
            /** @var \Illuminate\Http\Response|\Laravel\Lumen\Http\ResponseFactory $cached */
            $cached = $this->cache->get($cache_key);

            return $cached;
        } else {
            $this->validate($request, [
                'sources_urls'   => 'required|string|between:11,4096',
                'format'         => 'string|between:3,64',
                'version'        => 'string|between:1,8',
                'excluded_hosts' => 'string|between:1,4096',
                'limit'          => 'integer|min:1',
                'redirect_to'    => 'ip',
            ]);

            $format  = $request->get('format', 'routeros');
            $version = $request->get('version', config('app.version'));
            $limit   = $request->get('limit', (int) config('limits.result_entries', 0));
            $cache   = true;

            $hosts_parser = (new HostsParser)
                ->setCacheEnabled($cache)
                ->addComment(sprintf(
                    'Sources cache state: %s, *response* cache lifetime: %s',
                    $cache ? 'enabled' : 'disabled',
                    $this->cache_lifetime
                ))
                ->addExcludedHosts($request->get('excluded_hosts', config('defaults.excluded_hosts')))
                ->addSource($request->get('sources_urls'))
                ->setRedirectToAddress($request->get('redirect_to', config('defaults.redirect_ip')))
                ->requestSourcesResources();

            $result = response($hosts_parser->render($limit, config('defaults.script_ad_entries_comment'), $format))
                ->header('Content-Type', 'text/plain');

            $this->cache->put($cache_key, $result, Carbon::now()->addSeconds($this->cache_lifetime));

            return $result;
        }
    }
}
