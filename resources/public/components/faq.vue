<template>
    <div class="container">
        <h4>Как мне всё это дело прикрутить к моему MikroTik-у?</h4>
        <p>
            Более чем просто &mdash; необходимо добавить (<code>System</code> &rarr; <code>Scripts</code> &rarr; <code>Add
            New</code>) указанный выше скрипт, а так же добавить задание в планировщик (<code>System</code> &rarr;
            <code>Scheduler</code> &rarr; <code>Add New</code>) для его периодического запуска. Если если вы задали имя
            скрипта <code>AutoADBlock</code>, то в планировщике в поле <code>On Event</code> укажите: <code>/system
            script run AutoADBlock</code>. Права доступа: <code>[X] Read [X] Write [X] Policy [X] Test</code>.
        </p>

        <h4 class="mt-4">Выдача кэшируется?</h4>
        <p>В данный момент каждый запрашиваемый (<em>внешний</em>) ресурс кэшируется на <strong>{{ cacheLifetimeSec }}
            секунд</strong>. Всё остальное &mdash; обрабатывается в realtime.</p>

        <h4 class="mt-4">Какие ещё есть ограничения?</h4>
        <p>Ограничения хоть и носят больше формальный характер, но всё таки они есть:</p>
        <ul>
            <li>
                Максимальное количество внешних источников &mdash; <strong>{{ maxSourcesCount }}</strong> (<em>URI должен
                быть не более <code>{{ maxSourceUriLen}}</code> символов</em>);
            </li>
            <li>
                Максимальное количество исключений &mdash; <strong>{{ excludesLimit }}</strong>;
            </li>
            <li>
                Максимальный размер файла на источнике &mdash; <strong>{{ maxSourceSizeKb }}</strong> Кб;
            </li>
        </ul>

        <h4 class="mt-4">Применимо только к маршрутизаторам MikroTik (<em>RouterOS</em>)?</h4>
        <p>
            На данный момент &mdash; да. Но если потребуется дополнительный функционал &mdash; пишите <a
            :href="featureRequestLink" target="_blank">здесь</a>.
        </p>

        <h4 class="mt-4">У меня в таблице DNS есть нужные мне ресурсы. Как быть с ними?</h4>
        <p>С ними ничего не произойдет, так как импортируемые хосты "помечаются" определенным комментарием. Как
            следствие &mdash; они останутся буз изменений.</p>

        <h4 class="mt-4">Откуда источники хостов?</h4>
        <p>Используются открытые и обновляемые источники, указанные выше. Более того, вы можете указать свои источники
            (записи в которых имеют формат "<code>%ip_address% %host_name%</code>") доступные "извне" по протоколам:
            <code>http</code> и <code>https</code>. Разместить свой источник можете, например, на <a
                href="https://gist.github.com/" target="_blank">github gist</a> и указав ссылку на него (использовать
            <strong>RAW</strong> и только, кнопка на "raw" в правом-верхнем углу страницы gist-а).
        </p>

        <h4 class="mt-4">Запускаю указанный выше скрипт и ничего не происходит. Что делать?</h4>
        <p>Попробуйте выполнить в консоли <code>/system script print from=%имя_скрипта%</code> и проанализировать вывод.
            Работоспособность скрипта была протестирована на <code>RouterOS v6.30.2</code>.</p>

        <h4 class="mt-4">Я указал свой источник, но он не обрабатывается. Почему?</h4>
        <p>Указанный вами адрес должен отвечать кодом <strong>2xx</strong> (<em>или 3xx &mdash; но не больше <u>двух</u>
            редиректов</em>). Если при соблюдении этих условий он всё равно не обрабатывается, пожалуйста, напиши об
            этом <a :href="bugReportLink" target="_blank">вот тут</a>.
        </p>

        <h4 class="mt-4">Я не хочу, чтоб кто-то имел возможность выполнять произвольный код на моих маршрутизаторах. Но
            идея мне
            нравится. Что мне делать?</h4>
        <p>Данный "сервис" распространяется под лицензией MIT и исходники <a
            :href="projectLink" target="_blank">находятся в общем доступе</a>. Тебе остается только скачать, настроить
            и запустить его на своем ресурсе подконтрольном только тебе.</p>
    </div>
</template>

<script>
    /* global module */
    module.exports = {
        props: {
            cacheLifetimeSec: {
                default: 7200,
                type: Number
            },
            maxSourcesCount: {
                default: 8,
                type: Number
            },
            maxSourceUriLen: {
                default: 256,
                type: Number
            },
            excludesLimit: {
                default: 32,
                type: Number
            },
            maxSourceSizeKb: {
                default: 2048,
                type: Number
            },
            featureRequestLink: {
                default: 'https://github.com/tarampampam/mikrotik-hosts-parser/issues/new?template=feature_request.md',
                type: String
            },
            bugReportLink: {
                default: 'https://github.com/tarampampam/mikrotik-hosts-parser/issues/new?template=bug_report.md',
                type: String
            },
            projectLink: {
                default: 'https://github.com/tarampampam/mikrotik-hosts-parser',
                type: String
            },
        }
    }
</script>
