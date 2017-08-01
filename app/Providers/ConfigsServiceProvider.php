<?php

namespace App\Providers;

use Illuminate\Support\ServiceProvider;
use Laravel\Lumen\Application;

/**
 * Class ConfigsServiceProvider
 *
 * Загружаем все конфиги, описанные в директории ./config
 */
class ConfigsServiceProvider extends ServiceProvider
{
    /**
     * Регистрация сервис-провайдера.
     */
    public function register()
    {
        $app = $this->getApp();

        foreach (glob($app->getConfigurationPath() . '*.php', GLOB_ERR) as $path) {
            $app->configure(basename($path, '.php'));
        }
    }

    /**
     * Возвращает инстанс приложения.
     *
     * @return Application
     */
    protected function getApp()
    {
        return app(Application::class);
    }
}
