<?php
/**
 * Generate MikroTik static DNS AD-Block script
 *
 * @author    Samoylov Nikolay
 * @project   Hosts Parser and script 4 MikroTik generator
 * @copyright 2015 <github.com/tarampampam>
 * @license   MIT <http://opensource.org/licenses/MIT>
 * @github    <https://github.com/tarampampam/mikrotik-hosts-parser>
 * @version   0.0.4
 * @depends   PHP 5.x + curl
 */

class HostsParser {
  // Variable for calculate prepare + render time
  private $time_start = 0;
  
  // Array for store all hosts data
  private $hosts = array();
  
  // Array for sources links
  private $sources_list = array();
  
  // Array for hosts exceptions
  private $exceptions = array();
  
  // String with IP for redirection
  private $redirect_to = '127.0.0.1';
  
  // Limit total entries limit
  private $limit = 0;
  
  // Checking regex-s
  private $regex = array(
    'ip' => '(([01][0-9][0-9]\.|2[0-4][0-9]\.|[0-9][0-9]\.|25[0-5]\.|[0-9]\.)([01][0-9][0-9]\.|2[0-4][0-9]\.|[0-9][0-9]\.|25[0-5]\.|[0-9]\.)([01][0-9][0-9]\.|2[0-4][0-9]\.|[0-9][0-9]\.|25[0-5]\.|[0-9]\.)([01][0-9][0-9]|2[0-4][0-9]|25[0-5]|[0-9][0-9]|[0-9]))',
    'domain' => '((([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9]))',
    'url'    => '(http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?'
  );
  
  // User-agent for cURL
  private $useragent = 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.32 (KHTML, like Gecko) Chrome/36.0.2026.47 Safari/537.32';
  
  // Debug flag, show extended output
  public $debug = false;
  
  // cURL settings
  public $curl = array(
    'max_source_size' => 2097152, // 2097152 = 2 MiB
    'content_type' => 'text/plain',
  );
  
  // Cache settings
  public $cache = array(
    'enabled' => true,
    'path'    => '/cache',
    'expire'  => 15, // In seconds
  );
  
  /**
   * Class constructor
   *
   * Call '$HostsParser = new HostsParser()' at first
   */
  public function __construct() {
    // Setup start time
    $this->time_start = microtime(true);
  }
  
  /**
   * Output any string as 'log' message
   *
   * Call example: $this->log('Debug data in variable "$var": '.var_export($var, true));
   *
   * @param (string) Message
   */
  private function log($msg) {
    $msg = str_replace(array("  ", "\n", "\r", "\t"), "", $msg);
    $msg = str_replace(array(",)", ", )"), ")", $msg);
    $msg = str_replace("array (", "array(", $msg);
    echo('## '.$msg."\n");
  }
  
  /**
   * Normalize new line characters
   *
   * Windows (CRLF), Unix (LF), Mac (CR) format to Unix (LF) format + remove a lot of empty lines
   *
   * @param (string) String
   * @return (string) Normalized string
   */
  private function normalizeLewLines($str) {
    return preg_replace("/\n{2,}/", "\n\n", str_replace(array("\r\n", "\r"), "\n", $str));
  }
  
  /**
   * Get remote file content
   *
   * Use PHP extension named 'curl'. For install this extension you can exec in your server:
   *   $ sudo apt-get install php5-curl
   *   $ sudo /etc/init.d/apache2 restart
   *
   * @param (string) URL
   * @return (array) A lot of data, include headers and other info
   */
  private function curl($url) {
    $result = array('data' => '', 'success' => false);
    $cache_filename = $this->cache['path'].'/'.md5($url.$_SERVER['SERVER_NAME'].$_SERVER['SERVER_SOFTWARE']).'.txt';
    try {
      if(!function_exists('curl_init')) {
        if($this->debug) {
          $this->log('[Error] PHP::curl extension is not accessible');
        }
      } else {
        // Load from cache
        if($this->cache['enabled']) {
          if(
            file_exists($cache_filename) &&
            is_readable($cache_filename) &&
            (filemtime($cache_filename) > (time() - $this->cache['expire']))
          ) {
            $result['cache'] = true;
            $result['success'] = true;
            $result['data'] = file_get_contents($cache_filename);
            if($this->debug) {
              $this->log('[Cache] LOADED from cache: '.$url.' <-- '.$cache_filename);
            }
            return $result;
          }
        }
        $cURL = curl_init();
        if($cURL === false) {
          $result['data'] = 'Failed to initialize cURL';
          return $result;
        }

        // Get headers before
        curl_setopt_array($cURL, array(
          CURLOPT_URL            => $url,
          CURLOPT_RETURNTRANSFER => true,  // return web page
          CURLOPT_HEADER         => true,  // add headers to output
          CURLOPT_NOBODY         => true,  // NO BODY IN ANSWER
          CURLOPT_FOLLOWLOCATION => true,  // follow redirects
          CURLOPT_USERAGENT      => $this->useragent,
          CURLOPT_CONNECTTIMEOUT => 5,    // timeout on connect
          CURLOPT_TIMEOUT        => 5,    // timeout on response
          CURLOPT_MAXREDIRS      => 2,     // stop after N redirects 
          CURLOPT_SSL_VERIFYHOST => false, // Disabled SSL Cert checks
          CURLOPT_SSL_VERIFYPEER => false, // Disabled SSL Cert checks
        ));
        
        $result['data'] = curl_exec($cURL);
        
        if($this->debug) {
          $this->log('[Info] File by URL '.$url.' have size '.curl_getinfo($cURL, CURLINFO_CONTENT_LENGTH_DOWNLOAD));
        }
        
        if(
          ($result['data'] !== false) &&
          (curl_getinfo($cURL, CURLINFO_CONTENT_LENGTH_DOWNLOAD) <= $this->curl['max_source_size']) &&
          (strpos(strtolower(curl_getinfo($cURL, CURLINFO_CONTENT_TYPE)), $this->curl['content_type']) !== false)
        ) {
          $cURL = curl_init();
          curl_setopt_array($cURL, array(
            CURLOPT_URL            => $url,
            CURLOPT_RETURNTRANSFER => true,  // return web page
            CURLOPT_HEADER         => false, // add headers to output
            CURLOPT_FOLLOWLOCATION => true,  // follow redirects
            CURLOPT_USERAGENT      => $this->useragent,
            CURLOPT_CONNECTTIMEOUT => 15,    // timeout on connect
            CURLOPT_TIMEOUT        => 15,    // timeout on response
            CURLOPT_MAXREDIRS      => 2,     // stop after N redirects 
            CURLOPT_SSL_VERIFYHOST => false, // Disabled SSL Cert checks
            CURLOPT_SSL_VERIFYPEER => false, // Disabled SSL Cert checks
          ));
          
          $result['data'] = curl_exec($cURL);
          
          if($result['data'] === false) {
            $result['data'] = 'cURL error: Code '.curl_errno($cURL).' ('.curl_error($cURL).')';
            return $result;
          }
          
          $result['info'] = curl_getinfo($cURL);

          if($this->debug) {
            $this->log('Page info: '.var_export($result['info'], true));
          }
          
          curl_close($cURL);

          if(
            isset($result['info']) &&
            ($result['info']['http_code'] >= 200) &&
            ($result['info']['http_code'] < 400)
          ) {
            if(!empty($result['data'])) {
              // Save in cache
              if($this->cache['enabled']) {
                if(!is_dir($this->cache['path'])) {
                  mkdir($this->cache['path'], 0764, true);
                }
                file_put_contents($cache_filename, $result['data'], LOCK_EX);
                if($this->debug) {
                  $this->log('[Cache] SAVED in cache: '.$url.' --> '.$cache_filename);
                }
                $result['cache'] = false;
              }
              $result['success'] = true;
            }
          } else {
            $result['success'] = false;
          }
        } else {
          $result['success'] = false;
        }
      }
    } catch(Exception $e) {
      $result['data'] = 'cURL error: Code '.$e->getCode().' ('.$e->getMessage().')';
    }
    return $result;
  }
  
  /**
   * Validate IP address
   *
   * @param (string) IP address for test
   * @return (bool)  Validate result (true|false)
   */
  public function is_valid_ip($ip) {
    return preg_match('/^'.$this->regex['ip'].'$/', $ip);
  }
  
  /**
   * Validate domain name
   *
   * @param  (string) Domain name for test
   * @return (bool)   Validate result (true|false)
   */
  public function is_valid_domain($domain) {
    return preg_match('/^'.$this->regex['domain'].'$/', $domain);
  }
  
  /**
   * Validate URL path
   *
   * @param  (string) URL for test
   * @return (bool)   Validate result (true|false)
   */
  public function is_valid_url($url) {
    return preg_match('/'.$this->regex['url'].'/i', $url);
  }
  
  /**
   * Add source (URL) to external file
   *
   * @param (string|array) URL or array of URLs
   */
  public function sources_add($urls) {
    if(is_string($urls)) {
      $urls = array($urls);
    }
    if(is_array($urls) && !empty($urls)) {
      foreach($urls as $url) {
        if($this->is_valid_url($url) && !in_array($url, $this->sources_list)) {
          array_push($this->sources_list, $url);
        }
      }
    }
  }
  
  /**
   * Remove source (URL) from sources array
   *
   * @param (string|array) URL or array of URLs for remove
   */
  public function sources_remove($urls) {
    if(is_string($urls)) {
      $urls = array($urls);
    }
    if(is_array($urls) && !empty($urls)) {
      foreach($urls as $url) {
        if($this->is_valid_url($url) && in_array($url, $this->sources_list)) {
          $pos = array_search($url, $this->sources_list);
          if($pos !== false) {
            unset($this->sources_list[$pos]);
          }
        }
      }
    }
  }
  
  /**
   * Return sources array
   *
   * @return (array) Array
   */
  public function sources_get() {
    return $this->sources_list;
  }
  
  
  /**
   * Add host (domain) in array named '$exceptions' (exclude from result script)
   *
   * @param (string|array) Host name or array of host names
   */
  public function exceptions_add($hosts) {
    if(is_string($hosts)) {
      $hosts = array($hosts);
    }
    if(is_array($hosts) && !empty($hosts)) {
      foreach($hosts as $host) {
        if($this->is_valid_domain($host) && !in_array($host, $this->exceptions)) {
          array_push($this->exceptions, $host);
        }
      }
    }
  }
  
  /**
   * Remove host name from 'exceptions' array
   *
   * @param (string|array) Host name or array of host names
   */
  public function exceptions_remove($hosts) {
    if(is_string($hosts)) {
      $hosts = array($hosts);
    }
    if(is_array($hosts) && !empty($hosts)) {
      foreach($hosts as $host) {
        if($this->is_valid_domain($host) && in_array($host, $this->exceptions)) {
          $pos = array_search($host, $this->exceptions);
          if($pos !== false) {
            unset($this->exceptions[$pos]);
          }
        }
      }
    }
  }
  
  /**
   * Return 'exceptions' array
   *
   * @return (array) Array
   */
  public function exceptions_get() {
    return $this->exceptions;
  }
  
  
  /**
   * Add route rule to hosts array. Rule must be in format '127.0.0.1 localhost'
   *
   * @param (string|array) Rule or array of rules
   */
  public function routes_add($routes) {
    if(is_string($routes)) {
      $routes = array($routes);
    }
    if(is_array($routes) && !empty($routes)) {
      foreach($routes as $route) {
        $route = explode(' ', preg_replace('/[^a-zа-яё0-9\-\s\.\*]/i', '', trim(preg_replace('/\s+/', ' ', $route))));
        if(
          isset($route[0]) && !empty($route[0]) && $this->is_valid_ip($route[0]) &&
          isset($route[1]) && !empty($route[1]) && $this->is_valid_domain($route[1])
        ) {
          $this->hosts[$route[1]] = $route[0];
        }
      }
    }
  }
  
  /**
   * Remove route rule by host name from hosts array
   *
   * @param (string|array) Host name or array of host names
   */
  public function routes_remove($domains) {
    if(is_string($domains)) {
      $domains = array($domains);
    }
    if(is_array($domains) && !empty($domains)) {
      foreach($domains as $domain) {
        $domain = preg_replace('/[^a-zа-яё0-9\-\.\*]/i', '', $domain);
        if(!empty($this->hosts[$domain])) {
          unset($this->hosts[$domain]);
        }
      }
    }
  }
  
  /**
   * Return just user-defined route rules
   *
   * @return (array) Array of route rules
   */
  public function routes_get() {
    $routes = array();
    foreach($this->hosts as $key => $val) {
      if($this->is_valid_domain($key) && $this->is_valid_ip($val)) {
        $routes[$key] = $val;
      }
    }
    return $routes;
  }
  
  /**
   * Setup IP address for redirect
   *
   * @param (string) IP address
   */
  public function redirect_set($ip) {
    if(is_string($ip) && in_array($ip, array('localhost', 'loopback'))) {
      $ip = '127.0.0.1';
    }
    if(is_string($ip) && !empty($ip) && $this->is_valid_ip($ip)) {
      $this->redirect_to = $ip;
    }
  }
  
  /**
   * Return IP address for redirect
   *
   * @return (string) IP address
   */
  public function redirect_get() {
    return $this->redirect_to;
  }
  
  /**
   * Setup hosts limit for result script
   *
   * @param (string|numeric) Limit value
   */
  public function limit_set($limit) {
    if(is_string($limit) && in_array($limit, array('off', 'disable', 'unlimit'))) {
      $limit = 0;
    }
    if(is_numeric($limit) && $limit >= 0) {
      $this->limit = intval($limit, 10);
    }
  }
  
  /**
   * Return hosts limit for result script
   *
   * @return (numeric) Limit
   */
  public function limit_get() {
    return $this->limit;
  }
  
  /**
   * Return source hosts file data (using '$this->curl()' function)
   *
   * @param (string) URL string
   * @return (string|bool) File content of False if error
   */
  private function get_source_by_url($url) {
    if(is_string($url) && $this->is_valid_url($url)) {
      $hosts_data = $this->curl($url);
      if($hosts_data['success']) {
        return $this->normalizeLewLines($hosts_data['data']);
      } else {
        if($this->debug) {
          $this->log('[Error] File "'.$url.'" return error');
        }
        return false;
      }
    }
  }
  
  /**
   * Parse hosts file and store returned and valid in '$this->hosts'
   *
   * @param (string) Hosts file data
   */
  private function parse_hosts_file_data($hosts_data) {
    $hosts_raw = array();
    preg_match_all('#^'.$this->regex['ip'].'[\s\t]+'.$this->regex['domain'].'$#m', $hosts_data, $hosts_raw, PREG_SET_ORDER);
    foreach($hosts_raw as $hosts_entry) {
    //$ip   = $hosts_entry[1];
      $host = $hosts_entry[6];
      if(!in_array($host, $this->exceptions)) {
        array_push($this->hosts, $host);
      }
    }
  }
  
  /**
   * Render result MikroTik script
   *
   * @return (string) Script source
   */
  public function render() {
    //sort($this->hosts, SORT_STRING); // Make entries sorting

    // Get sources data and store in $this->hosts
    foreach($this->sources_list as $source) {
      $this->parse_hosts_file_data($this->get_source_by_url($source));
    }
    
    // Remove any duplicates
    $this->hosts = array_unique($this->hosts, SORT_STRING);
    
    $result = '## Script generated '.date('m/d/Y \a\t H:i:s', time()).' '.
              '('.count($this->hosts).' entries'.(($this->limit > 0) ? ', limited to '.$this->limit : '').') '.
              'for '.round((microtime(true) - $this->time_start), 3).' sec.'."\n";
    
    if(!empty($this->sources_list)) {
      $result .= "##\n".'## Hosts list sources:'."\n";
      foreach($this->sources_list as $list_item) {
        $result .= '## > '.$list_item."\n";
      }
    }
    
    if(!empty($this->exceptions)) {
      $result .= "##\n".'## Exception hosts:'."\n";
      foreach($this->exceptions as $host) {
        $result .= '## > '.$host."\n";
      }
    }
    
    $result .= "\n".'/ip dns static'."\n";
    
    // Limit hosts count
    if(($this->limit > 0) && (count($this->hosts) > $this->limit)) {
      $this->hosts = array_slice($this->hosts, 0, $this->limit);
    }
    
    foreach($this->hosts as $key => $value) {
      if(!is_numeric($key) && $this->is_valid_ip($value) && $this->is_valid_domain($key)) {
        $result .= 'add address='.$value.' name='.$key."\n";
      } elseif(!empty($key)) {
        $result .= 'add address='.$this->redirect_to.' name='.$value."\n";
      }
    }
    return $result;
  }
}

