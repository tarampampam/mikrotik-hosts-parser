<?php

/**
 * Возвращает настройки ограничений.
 */

return [

    // Количество источников, которые может указать пользователь
    'user_sources' => (int) env('LIMIT_USER_SOURCES', 8),

    // Количество итоговых записей, возвращаемых в генерируемом скрипте; 0 - без ограничений
    'result_entries' => (int) env('LIMIT_RESULT_ENTRIES', 0),

    'source' => [
        'cache' => [
            // Время жизни кэша ответа от источника (в секундах)
            'lifetime' => (int) env('LIMIT_SOURCE_CACHE_LIFETIME', 7200),
        ],
    ],

    // Максимальное количество возможных редиректов
    'max_redirects_count' => (int) env('LIMIT_MAX_REDIRECTS', 2),

    // Максимальное количество исключенных пользовательских хостов
    'excluded_hosts' => (int) env('LIMIT_EXCLUDED_HOSTS', 32),

    // Максимальный размер файла-источника (в КилоБайтах)
    'source_file_size' => (int) env('LIMIT_SOURCE_LIFE_SIZE', 2048),

    // Максимальная длинна URI до файла-источника
    'source_uri_length' => (int) env('LIMIT_SOURCE_URI_LENGTH', 256),

    // Протоколы, по которым можно стучаться к файлам-источникам
    'sources_protocols' => explode(',', env('EXCLUDED_SOURCES_PROTOCOLS', 'http,https,ftp')),

];
