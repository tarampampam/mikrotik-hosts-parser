<?php

namespace Tests;

class ExampleTest extends TestCase
{
    /**
     * Test of index page.
     *
     * @return void
     */
    public function testIndexPage()
    {
        $response = $this->get('/');

        $response->assertResponseStatus(200);
    }

    /**
     * Test of script source page.
     *
     * @return void
     */
    public function testScriptSourcePage()
    {
        $response = $this->get('/script/source?format=routeros&version=2.0.2&redirect_to=127.0.0.2&limit=666&sources_urls=https%3A%2F%2Fcdn.rawgit.com%2Ftarampampam%2Fstatic%2Fmaster%2Fhosts%2Fblock_shit.txt&excluded_hosts=localhost');

        $response->assertResponseOk();
        $this->assertContains('Script generated', $response->response->getContent());

        $response->assertResponseStatus(200);
    }
}
