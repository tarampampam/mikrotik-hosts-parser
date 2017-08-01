@extends('layouts.app')

@section('html_header')
    <link type="text/css" href="{{ url('/components/highlightjs/styles/default.css') }}" rel="stylesheet"/>
    <link type="text/css" href="{{ url('/components/highlightjs/styles/github-gist.css') }}" rel="stylesheet"/>
@endsection

@section('main_content')
    <div class="headline">
        <h1 class="logo">
            <span class="logo"><i class="fa fa-flash"></i></span>
            MikroTik StopAD
            <small>Сделаем рекламы чуть меньше</small>
        </h1>
    </div> <!-- /demo-headline -->

    <div class="container">
        <h3>Что это такое?</h3>
        <p>Это сервис, который выполняет одну-единственную функцию &mdash; генерирует скрипт для маршрутизаторов
            <strong>MikroTik</strong> (<em>Router OS</em>), импортируя который, производится блокировка "рекламных"
            (<em>и не только</em>) доменов на основании как публичных, так и указанных вами хост-листов.</p>
    </div>

    <div class="container">
        <h4>Как блокируются?</h4>
        <p>До безобразия просто &mdash; все <abbr title="Domain Name System — система доменных имён">DNS</abbr> запросы
            доменов, которые проходят через маршрутизатор (<em>если он указан первым DNS сервером</em>) в случае
            соответствия с "рекламным" &mdash; перенаправляются, например, на <code>127.0.0.1</code> (<em>локальный
                хост</em>).</p>
    </div>

    <div class="container">
        <blockquote>
            <h6>У меня не работает / перестало работать</h6>
            <p class="small">Если ранее всё работало как надо, и <em>внезапно</em> перестало &mdash; то вероятнее всего
                это произошло по причине обновления логики работы скрипта. Поддержки старых версий нет, поэтому &mdash;
                просто обнови скрипт в маршрутизаторе (<em>возьми его обновленную версию прямо с этой страницы, его
                    исходник чуть ниже</em>). Во всех остальных случаях &mdash; есть смысл попытаться оставить
                комментарий у <a href="http://blog.kplus.pro/mikrotik/remove-a-lot-of-ad-using-mikrotik.html"
                                 target="_blank">этого поста в уютном блоге</a> с подробным описанием что ты пытался
                сделать, каким образом, и что получилось в итоге.</p>
        </blockquote>
    </div>

    <hr class="rainbow"/>

    <div class="form-group clearfix">

        <h5>Источники</h5>
        <div class="row">
            @isset($sources)
                <div class="col-xs-6">
                    @foreach($sources as $source)
                        <label class="checkbox">
                            <input type="checkbox" data-toggle="checkbox" class="custom-checkbox source-checkbox"
                                   data-url="{{ $source['uri'] }}" {{ (isset($source['checked']) && $source['checked']
                                   === true) ? 'checked' : '' }}/>
                            <span class="icons"><span class="icon-unchecked"></span><span
                                        class="icon-checked"></span></span>
                            {{ $source['title'] or 'Нет описания' }} &nbsp;<a href="{{ $source['uri'] }}"
                                                                              target="_blank"><i class="fa
                                                                         fa-external-link"></i></a>
                        </label>
                    @endforeach
                </div>
            @endisset
            <div class="col-xs-6">
                <p class="small">При необходимости можете указать свои источники, по одному на строку <em>(общий
                        лимит составляет <strong>{{ $user_sources_limit or '&#8734;' }}</strong>
                        источников)</em>:</p>
                <textarea class="form-control user-defined-sources" placeholder="http://winhelp2002.mvps.org/hosts.txt"
                          rows="7"></textarea>
            </div>
        </div>


        <div class="row">
            <div class="col-xs-6">
                <h6>Адрес перенаправления</h6>
                <p class="small">Укажите IP (v4) адрес, куда перенаправлять запросы:</p>
                <div class="form-group">
                    <input type="text" value="" placeholder="{{ $default_redirect_ip or '127.0.0.1' }}"
                           class="form-control redirect-to-ip"/>
                </div>

                <h6>Лимит записей</h6>
                <p class="small">Укажите максимальное количество возвращаемых записей:</p>
                <div class="form-group">
                    <input type="text" value="" placeholder="{{ $result_entries_limit or '0' }}"
                           class="form-control result-limit"/>
                </div>
            </div>
            <div class="col-xs-6">
                <h6>Исключения</h6>
                <p class="small">Можете указать те хосты, которые необходимо исключить из итогового скрипта, одна строка
                    для одного хоста:</p>
                <div class="form-group">
                    <textarea class="form-control excluded-hosts" placeholder="adserver.yahoo.com"
                              rows="6">{{ $excluded_hosts or '' }}</textarea>
                </div>
            </div>
        </div>
    </div>

    <hr class="rainbow"/>

    <div class="form-group clearfix">
        <h5>Строка запроса</h5>
        <div class="row">
            <div class="col-xs-12">
                <pre class="script-uri hljs accesslog"></pre>
            </div>
        </div>
        <h6>Скрипт для маршрутизатора</h6>
        <div class="row">
            <div class="col-xs-12">
                <pre class="script-source hljs routeros">{{ $script_source or '' }}</pre>
            </div>
        </div>
    </div>

    <hr class="rainbow"/>

    <div class="container">
        <h6>Как мне всё это дело прикрутить к моему MikroTik-у?</h6>
        <p>
            Более чем просто &mdash; необходимо добавить (<code>System</code> &rarr; <code>Scripts</code> &rarr; <code>Add
                New</code>) указанный выше скрипт, а так же добавить задание в планировщик (<code>System</code> &rarr;
            <code>Scheduler</code> &rarr; <code>Add New</code>) для его периодического запуска. Если если вы задали имя
            скрипта <code>AutoADBlock</code>, то в планировщике в поле <code>On Event</code> укажите: <code>/system
                script run AutoADBlock</code>. Права доступа: <code>[X] Read [X] Write [X] Policy [X] Test</code>.
        </p>
    </div>
    <div class="container">
        <h6>Выдача кэшируется?</h6>
        <p>В данный момент каждый запрашиваемый (<em>внешний</em>) ресурс кэшируется на
            <strong>{{ $source_cache_lifetime or '&#8734;' }} секунд</strong>. Всё остальное &mdash; обрабатывается в
            реалтайме.</p>
    </div>
    <div class="container">
        <h6>Какие ещё есть ограничения?</h6>
        <p>Ограничения хоть и носят больше формальный характер, но всё таки они есть:</p>
        <ul>
            <li>Максимальное количество внешних источников &mdash;
                <strong>{{ $user_sources_limit or '&#8734;' }}</strong> (<em>URL должен быть не более
                    <code>{{ $source_uri_length or '&#8734;' }}</code> символов</em>);
            </li>
            <li>Максимальное количество исключений &mdash; <strong>{{ $excluded_hosts_limit or '&#8734;' }}</strong>;
            </li>
            <li>Максимальный размер файла на источнике &mdash;
                <strong>{{ $source_file_size_limit or '&#8734;' }}</strong> Кб. (<em>а так же поле
                    <code>content_type</code> должно соответствовать <code>text/plain</code></em>);
            </li>
        </ul>
    </div>
    <div class="container">
        <h6>Применимо только к маршрутизаторам MikroTik (<em>RouterOS</em>)?</h6>
        <p>
            На данный момент &mdash; да. Но если потребуется дополнительный функционал &mdash; пишите <a
                    href="{{ config('contacts.repository.issues.new') }}" target="_blank">здесь</a>.</p>
    </div>
    <div class="container">
        <h6>У меня в таблице DNS есть нужные мне ресурсы. Как быть с ними?</h6>
        <p>Так как перед импортированием скрипта потребуется уничтожить все имеющиеся маршруты. Вы можете задать
            служебные (<em>ваши</em>) маршруты предварительно и они будут включены в итоговый скрипт. Таким образом ваши
            маршруты будут сохранены.</p>
    </div>
    <div class="container">
        <h6>Откуда источники хостов?</h6>
        <p>Мы используем открытые и обновляемые источники, указанные выше. Более того, вы можете указать свои источники
            (записи в которых имеют формат "<code>%ip_address% %host_name%</code>") доступные "извне" по протоколам:
            @if (isset($sources_protocols) && is_array($sources_protocols) && !empty($sources_protocols))
                <code>{!! implode('</code>, <code>', $sources_protocols) !!}</code>
            @else
                <code>Не указаны в настройках</code>
            @endif
            .
        </p>
    </div>
    <div class="container">
        <h6>Запускаю указанный выше скрипт и ничего не происходит. Что делать?</h6>
        <p>Попробуйте выполнить в консоли <code>/system script print from=%имя_скрипта%</code> и проанализировать вывод.
            Работоспособность скрипта была протестирована на <code>RouterOS v6.30.2</code>.</p>
    </div>
    <div class="container">
        <h6>Я указал свой источник, но он не обрабатывается. Почему?</h6>
        <p>Указанный вами адрес должен отвечать кодом <strong>2xx</strong> (<em>или 3xx &mdash; но не больше <u>двух</u>
                редиректов</em>). Если при соблюдении этих условий он всё равно не обрабатывается, пожалуйста, напиши об
            этом <a href="{{ config('contacts.repository.issues.new') }}" target="_blank">вот тут</a>.</p>
    </div>
    <div class="container">
        <h6>Я не хочу, чтоб кто-то имел возможность выполнять произвольный код на моих маршрутизаторах. Но идея мне
            нравится. Что мне делать?</h6>
        <p>Данный "сервис" распространяется под лицензией MIT и исходники парсера <a
                    href="{{ config('contacts.repository.uri') }}" target="_blank">находятся в общем
                доступе</a>. Тебе остается только скачать, настроить и запустить его на своем ресурсе подконтрольном
            только тебе.</p>
    </div>
@endsection

@section('inline_scripts')
    <script type="text/javascript" src="{{ url('/components/highlightjs/highlight.pack.min.js') }}"></script>
    <script type="text/javascript" src="{{ url('/assets/js/index.js') }}"></script>
@endsection