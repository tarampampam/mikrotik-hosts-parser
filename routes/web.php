<?php

/*
|--------------------------------------------------------------------------
| Application Routes
|--------------------------------------------------------------------------
|
| Here is where you can register all of the routes for an application.
| It is a breeze. Simply tell Lumen the URIs it should respond to
| and give it the Closure to call when that URI is requested.
|
*/

$app->get('/', [
    'as'   => 'web.index',
    'uses' => 'WebController@index',
]);

$app->get('/script/source', [
    'as'   => 'script.source',
    'uses' => 'ScriptController@source',
]);
