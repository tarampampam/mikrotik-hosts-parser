<?php

namespace App\Services\HostsParser;

use Exception;
use Carbon\Carbon;
use GuzzleHttp\Client;
use Illuminate\Support\Str;
use Psr\Http\Message\ResponseInterface;

/**
 * Class HostsParser.
 */
class HostsParser
{
    /**
     * Comments lines.
     *
     * @var string[]|array
     */
    protected $comments = [];

    /**
     * Hosts names.
     *
     * @var string[]|array
     */
    protected $hosts = [];

    /**
     * Array of sources URIs.
     *
     * @var string[]|array
     */
    protected $sources_list = [];

    /**
     * Array of excluded hosts names.
     *
     * @var string[]|array
     */
    protected $excluded_hosts = [];

    /**
     * Redirect IP address.
     *
     * @var string
     */
    protected $redirect_to = '127.0.0.1';

    /**
     * Cache enabled?
     *
     * @var bool
     */
    protected $cache_enabled = false;

    /**
     * Set "redirect to" address.
     *
     * @param string $address
     *
     * @return $this
     */
    public function setRedirectToAddress($address)
    {
        $this->redirect_to = (string) $address;

        return $this;
    }

    /**
     * Get "redirect to" address.
     *
     * @return string
     */
    public function getRedirectToAddress()
    {
        return $this->redirect_to;
    }

    /**
     * Cache is enabled?
     *
     * @return bool
     */
    public function getCacheEnabled()
    {
        return $this->cache_enabled;
    }

    /**
     * Set enabled (or disabled) cache state.
     *
     * @param bool $enabled
     *
     * @return $this
     */
    public function setCacheEnabled($enabled)
    {
        $this->cache_enabled = (bool) $enabled;

        return $this;
    }

    /**
     * Add source URI.
     *
     * @param string|array $source_uri
     *
     * @return $this
     */
    public function addSource($source_uri)
    {
        if (is_string($source_uri) && Str::contains($source_uri, ',')) {
            $source_uri = explode(',', $source_uri);
        }

        foreach ((array) $source_uri as $uri) {
            $uri = urldecode($uri);
            if ($this->isValidUri($uri) && ! in_array($uri, $this->sources_list)) {
                array_push($this->sources_list, $uri);
            }
        }

        return $this;
    }

    /**
     * Validate URI string.
     *
     * @param string $uri
     *
     * @return bool|mixed
     */
    public function isValidUri($uri)
    {
        return $this->validate($uri, 'url');
    }

    /**
     * Add hostname to the excluded hosts stack.
     *
     * @param string[]|string $hosts_names
     *
     * @return $this
     */
    public function addExcludedHosts($hosts_names)
    {
        if (is_string($hosts_names) && Str::contains($hosts_names, ',')) {
            $hosts_names = explode(',', $hosts_names);
        }

        foreach ((array) $hosts_names as $hostname) {
            $hostname = urldecode($hostname);
            if ($this->isValidHostname($hostname) && ! in_array($hostname, $this->excluded_hosts)) {
                array_push($this->excluded_hosts, $hostname);
            }
        }

        return $this;
    }

    /**
     * Validate hostname.
     *
     * @param string $hostname
     *
     * @return bool
     */
    public function isValidHostname($hostname)
    {
        static $regexp = '((([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]'
            . '*[A-Za-z0-9]))';

        if (is_string($hostname) && ! empty($hostname)) {
            if ((bool) preg_match('/^' . $regexp . '$/', $hostname)) {
                return true;
            }
        }

        return false;
    }

    /**
     * Load external hosts lists contend and add to the hosts stack.
     *
     * @return $this
     */
    public function requestSourcesResources()
    {
        foreach ((array) $this->getSourcesList() as $source_uri) {
            try {
                $response = $this->makeRequest($source_uri);
                $this->addHostsNames($this->extractHostsNamesFromHostsFile($response));
            } catch (Exception $e) {
                $message = 'Error :: ' . $source_uri . ' :: %s';
                if (($previous = $e->getPrevious()) && $previous instanceof Exception) {
                    $this->addComment(sprintf($message, $previous->getMessage()));
                } else {
                    $this->addComment(sprintf($message, $e->getMessage()));
                }
            }
        }

        return $this;
    }

    /**
     * Get sources URIs list.
     *
     * @return string[]|array
     */
    public function getSourcesList()
    {
        return $this->sources_list;
    }

    /**
     * Make a request to the source.
     *
     * @param string $uri
     *
     * @throws Exception
     *
     * @return bool|string
     */
    public function makeRequest($uri)
    {
        if ($this->isValidUri($uri)) {
            $cache_key = md5($uri);
            if ($this->cache_enabled && $this->getCacheRepository()->has($cache_key)) {
                $this->addComment(sprintf('Cache: "%s" loaded from cache', $uri));

                return $this->getCacheRepository()
                            ->get($cache_key);
            } else {
                $http_client = new Client($this->getDefaultHttpClientOptions());
                $response    = $http_client->request('get', $uri);
                $content     = $this->normalizeNewLineCodes($response->getBody()->getContents());
                if (is_string($content) && ! empty($content)) {
                    if ($this->cache_enabled) {
                        $this->getCacheRepository()
                             ->put($cache_key, $content, (int) config('limits.source.cache.lifetime', 600));
                    }

                    return $content;
                }
            }
        }

        return false;
    }

    /**
     * Get user-agent string fo a work.
     *
     * @return string
     */
    public function getUserAgent()
    {
        return 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.32 (KHTML, like Gecko) Chrome/36.0.2026.47 '
            . 'Safari/537.32';
    }

    /**
     * Normalize new line characters.
     *
     * @param string $string
     *
     * @return null|string
     */
    public function normalizeNewLineCodes($string)
    {
        if (is_scalar($string) && ! empty($string)) {
            return (string) preg_replace(
                "/\n{2,}/",
                "\n\n",
                str_replace(["\r\n", "\n\r", "\r"], "\n", (string) $string)
            );
        }
    }

    /**
     * Add hostname to the hosts stack.
     *
     * @param string[]|string $hosts_names
     *
     * @return $this
     */
    public function addHostsNames($hosts_names)
    {
        foreach ((array) $hosts_names as $hostname) {
            if ($this->isValidHostname($hostname) && ! in_array($hostname, $this->hosts)) {
                array_push($this->hosts, $hostname);
            }
        }

        return $this;
    }

    /**
     * Extract hosts names from raw hosts file content.
     *
     * @param string $raw_content
     *
     * @return array
     */
    public function extractHostsNamesFromHostsFile($raw_content)
    {
        $result = [];

        if (is_string($raw_content) && ! empty($raw_content)) {
            $pattern = '~([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})[\s\t]+(?P<domain_names>[^\s\\\/]+)~m';
            preg_match_all($pattern, $raw_content, $matches);

            if (isset($matches['domain_names']) && ! empty($matches['domain_names'])) {
                $result = array_filter(array_map(function ($hostname) {
                    $hostname = Str::lower(trim($hostname));

                    return $this->isValidHostname($hostname)
                        ? $hostname
                        : null;
                }, $matches['domain_names']));
            }
        }

        return $result;
    }

    /**
     * Add comment message.
     *
     * @param string[]|string $messages
     *
     * @return $this
     */
    public function addComment($messages)
    {
        foreach ((array) $messages as $message) {
            array_push($this->comments, (string) $message);
        }

        return $this;
    }

    /**
     * Validate IP address.
     *
     * @param string $ip
     *
     * @return bool|mixed
     */
    public function isValidIPv4($ip)
    {
        return $this->validate($ip, 'ip');
    }

    /**
     * Render result script.
     *
     * @param mixed $limit
     * @param mixed $entry_comment
     * @param mixed $format
     *
     * @return string
     */
    public function render($limit = 0, $entry_comment = 'ADBlock', $format = 'routeros')
    {
        $hosts = $this->getHostsNames();

        $this->addComment([
            sprintf('Script generated %s', Carbon::now()->toDateTimeString()),
            sprintf('Script format: "%s"', $format),
            null,
        ]);

        $this->addComment(array_merge(['Hosts list sources:'], array_map(function ($source_uri) {
            return sprintf('> %s', $source_uri);
        }, $this->getSourcesList()), [null]));

        $this->addComment(array_merge(['Excluded hosts sources:'], array_map(function ($host_name) use (&$hosts) {
            // Remove excluded hosts
            if (($key = array_search($host_name, $hosts)) !== false) {
                unset($hosts[$key]);
            }

            return sprintf('> %s', $host_name);
        }, $this->getExcludedHosts())));

        if ((int) $limit > 0) {
            if (count($hosts) > $limit) {
                $this->addComment([null, sprintf('%d entries limited to %d', count($hosts), $limit)]);
                $hosts = array_slice($hosts, 0, $limit);
            }
        }

        $result = '';

        // Append comments lines to result
        foreach ($this->comments as $comment) {
            $result .= sprintf("## %s\n", $comment);
        }

        if (count($hosts) >= 1) {
            $result .= "\n/ip dns static\n";

            foreach ($hosts as $host) {
                $result .= sprintf(
                    "add address=%s name=%s comment=%s\n",
                    $this->redirect_to,
                    $host,
                    str_replace(' ', '', $entry_comment)
                );
            }
        }

        return $result;
    }

    /**
     * Get hosts names stack.
     *
     * @return array|string[]
     */
    public function getHostsNames()
    {
        return $this->hosts;
    }

    /**
     * Get excluded hosts list.
     *
     * @return array|string[]
     */
    public function getExcludedHosts()
    {
        return $this->excluded_hosts;
    }

    /**
     * @param string          $value
     * @param string|string[] $rules
     *
     * @return bool
     */
    protected function validate($value, $rules)
    {
        static $stack = [];

        if (is_array($rules) && ! empty($rules)) {
            $rules = implode('|', $rules);
        }

        if (is_string($value) && ! empty($value) && is_string($rules) && ! empty($rules)) {
            if (! isset($stack[$value])) {
                $stack[$value] = ! $this->getValidationFactory()
                                       ->make(['value' => $value], ['value' => 'required|' . $rules])
                                       ->fails();
            }

            return (bool) $stack[$value];
        }

        return false;
    }

    /**
     * Get a validation factory instance.
     *
     * @return \Illuminate\Contracts\Validation\Factory
     */
    protected function getValidationFactory()
    {
        return app('validator');
    }

    /**
     * Get a cache repository instance.
     *
     * @return \Illuminate\Cache\Repository
     */
    protected function getCacheRepository()
    {
        static $instance = null;

        if (is_null($instance)) {
            $instance = app('cache');
        }

        return $instance;
    }

    /**
     * Get default HTTP client options.
     *
     * @return array
     */
    protected function getDefaultHttpClientOptions()
    {
        return [
            'timeout'         => $this->getHttpClientTimeout(),
            'connect_timeout' => $this->getHttpClientTimeout(),

            'allow_redirects' => [
                'max'       => config('limits.max_redirects_count', 5),
                'protocols' => config('limits.sources_protocols', ['http', 'https']),
            ],

            'headers'     => [
                'User-Agent' => $this->getUserAgent(),
            ],

            // Cancel download, if found header with value in content-length more then we have in config
            'on_headers'  => function (ResponseInterface $response) {
                $content_length = $response->getHeaderLine('Content-Length');
                if (! empty($content_length) && is_scalar($content_length)) {
                    if ((intval($content_length, 10) / 1024) > $this->getDownloadFileSizeLimit()) {
                        throw new Exception('The file is too big (detected by header "Content-Length")');
                    }
                }

                $content_type = $response->getHeaderLine('Content-Type');
                if (! empty($content_type) && is_scalar($content_type)) {
                    if (! Str::contains(Str::lower((string) $content_type), 'text/plain')) {
                        throw new Exception(sprintf('Invalid content type header (%s)', $content_type));
                    }
                }
            },

            // Cancel download, if downloaded content size more then limit, declared in config
            'progress'    => function ($download_total, $downloaded_bytes) {
                if ((intval($downloaded_bytes, 10) / 1024) > $this->getDownloadFileSizeLimit()) {
                    throw new Exception('The file is too big (detected by "progress" callback)');
                }
            },

            // Set to false to disable throwing exceptions on an HTTP protocol errors (i.e., 4xx and 5xx responses)
            'http_errors' => true,

            'verify' => false,
        ];
    }

    /**
     * Get HTTP client timeout.
     *
     * @return int
     */
    protected function getHttpClientTimeout()
    {
        return 5;
    }

    /**
     * Get download file size limit (in kilobytes).
     *
     * @return int
     */
    protected function getDownloadFileSizeLimit()
    {
        static $limit = null;

        if (is_null($limit)) {
            $limit = (int) config('limits.source_file_size', 2048);
        }

        return $limit;
    }
}
