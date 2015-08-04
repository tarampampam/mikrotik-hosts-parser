<?php

## @author    Samoylov Nikolay
## @project   Hosts Parser 4 MikroTik
## @copyright 2015 <github.com/tarampampam>
## @license   MIT <http://opensource.org/licenses/MIT>
## @github    <https://github.com/tarampampam/mikrotik-hosts-parser>

define('BasePath', realpath(dirname(__FILE__)), true);

// Подключаем класс с парсером
$hostsparser_class_file = BasePath.'/hostsparser.class.php';
if(file_exists($hostsparser_class_file)){ require_once($hostsparser_class_file); }

// Создаем экземпляр класса
$HostsParser = new HostsParser();

// Исключительно для отладки и поиска ошибок
$HostsParser->debug = true;

// Настройка кэширования
$HostsParser->cache['path'] = BasePath.'/cache';
$HostsParser->cache['expire'] = 300; // В секундах
$HostsParser->cache['enabled'] = true;

// Настройки cURL
$HostsParser->curl['max_source_size'] = 2097152; // 2097152 = 2 MiB
$HostsParser->curl['content_type'] = 'text/plain';

// Настройки парсера
$HostsParser->routes_add(array('127.0.0.1 localhost', '192.168.1.1 router'));
$HostsParser->exceptions_add('localhost');
$HostsParser->redirect_set('127.0.0.1');
$HostsParser->limit_set(0);
$HostsParser->sources_add(array('http://adaway.org/hosts.txt', 'http://winhelp2002.mvps.org/hosts.txt'));

// Выводим результат
header('Content-Type: text/plain');
echo($HostsParser->render());
