<?php

namespace App\Http\Controllers;

use Laravel\Lumen\Routing\Controller as BaseController;

/**
 * Class AbstractController.
 */
abstract class AbstractController extends BaseController
{
    /**
     * @param $any
     *
     * @return string
     */
    protected function getHash($any)
    {
        return md5(serialize($any));
    }
}
