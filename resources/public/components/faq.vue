<template>
    <div class="container">
        <h5>Как мне всё это дело прикрутить к моему MikroTik-у?</h5>
        <p>
            Более чем просто &mdash; необходимо добавить (<code>System</code> &rarr; <code>Scripts</code> &rarr; <code>Add
            New</code>) указанный выше скрипт, а так же добавить задание в планировщик (<code>System</code> &rarr;
            <code>Scheduler</code> &rarr; <code>Add New</code>) для его периодического запуска. Если если вы задали имя
            скрипта <code>AutoADBlock</code>, то в планировщике в поле <code>On Event</code> укажите: <code>/system
            script run AutoADBlock</code>. Права доступа: <code>[X] Read [X] Write [X] Policy [X] Test</code>.
        </p>

        <h5 class="mt-4">Выдача кэшируется?</h5>
        <p>В данный момент каждый запрашиваемый (<em>внешний</em>) ресурс кэшируется на <strong>{{ cache_lifetime_sec }}
            секунд</strong>. Всё остальное &mdash; обрабатывается в realtime.</p>

        <h5 class="mt-4">Какие ещё есть ограничения?</h5>
        <p>Ограничения хоть и носят больше формальный характер, но всё таки они есть:</p>
        <ul>
            <li>
                Максимальное количество внешних источников &mdash; <strong>{{ max_sources_count }}</strong> (<em>URI должен
                быть не более <code>{{ max_source_uri_len}}</code> символов</em>);
            </li>
            <li>
                Максимальное количество исключений &mdash; <strong>{{ excludes_limit }}</strong>;
            </li>
            <li>
                Максимальный размер файла на источнике &mdash; <strong>{{ max_source_size_kb }}</strong> Кб;
            </li>
        </ul>

        <h5 class="mt-4">Применимо только к маршрутизаторам MikroTik (<em>RouterOS</em>)?</h5>
        <p>
            На данный момент &mdash; да. Но если потребуется дополнительный функционал &mdash; пишите <a
            :href="feature_request_link" target="_blank">здесь</a>.
        </p>

        <h5 class="mt-4">У меня в таблице DNS есть нужные мне ресурсы. Как быть с ними?</h5>
        <p>С ними ничего не произойдет, так как импортируемые хосты "помечаются" определенным комментарием. Как
            следствие &mdash; они останутся буз изменений.</p>

        <h5 class="mt-4">Откуда источники хостов?</h5>
        <p>Используются открытые и обновляемые источники, указанные выше. Более того, вы можете указать свои источники
            (записи в которых имеют формат "<code>%ip_address% %host_name%</code>") доступные "извне" по протоколам:
            <code>http</code> и <code>https</code>. Разместить свой источник можете, например, на <a
                href="https://gist.github.com/" target="_blank">github gist</a> и указав ссылку на него (использовать
            <strong>RAW</strong> и только, кнопка на "raw" в правом-верхнем углу страницы gist-а).
        </p>

        <h5 class="mt-4">Запускаю указанный выше скрипт и ничего не происходит. Что делать?</h5>
        <p>Попробуйте выполнить в консоли <code>/system script print from=%имя_скрипта%</code> и проанализировать вывод.
            Работоспособность скрипта была протестирована на <code>RouterOS v6.30.2</code>.</p>

        <h5 class="mt-4">Я указал свой источник, но он не обрабатывается. Почему?</h5>
        <p>Указанный вами адрес должен отвечать кодом <strong>2xx</strong> (<em>или 3xx &mdash; но не больше <u>двух</u>
            редиректов</em>). Если при соблюдении этих условий он всё равно не обрабатывается, пожалуйста, напиши об
            этом <a :href="bug_report_link" target="_blank">вот тут</a>.
        </p>

        <h5 class="mt-4">Я не хочу, чтоб кто-то имел возможность выполнять произвольный код на моих маршрутизаторах. Но
            идея мне
            нравится. Что мне делать?</h5>
        <p>Данный "сервис" распространяется под лицензией MIT и исходники <a
            :href="project_link" target="_blank">находятся в общем доступе</a>. Тебе остается только скачать, настроить
            и запустить его на своем ресурсе подконтрольном только тебе.</p>
    </div>
</template>

<script>
    /* global module */
    module.exports = {
        props: {
            cache_lifetime_sec: {
                default: 7200,
                type: Number
            },
            max_sources_count: {
                default: 8,
                type: Number
            },
            max_source_uri_len: {
                default: 256,
                type: Number
            },
            excludes_limit: {
                default: 32,
                type: Number
            },
            max_source_size_kb: {
                default: 2048,
                type: Number
            },
            feature_request_link: {
                default: 'https://github.com/tarampampam/mikrotik-hosts-parser/issues/new?template=feature_request.md',
                type: String
            },
            bug_report_link: {
                default: 'https://github.com/tarampampam/mikrotik-hosts-parser/issues/new?template=bug_report.md',
                type: String
            },
            project_link: {
                default: 'https://github.com/tarampampam/mikrotik-hosts-parser',
                type: String
            },
        }
    }
</script>

<style scoped>

</style>
