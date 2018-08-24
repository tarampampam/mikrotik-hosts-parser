    <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/1.11.3/jquery.min.js"
            integrity="sha256-rsPUGdUPBXgalvIj4YKJrrUlmLXbOb6Cp7cdxn1qeUc="
            crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/flat-ui/2.2.2/js/flat-ui.min.js"
            integrity="sha256-M8wr/v1TanHRGWD9MyiHRqwB0pzAUjjUVDyzq8MInY0="
            crossorigin="anonymous"></script>

    @if(isset($js_vars) && is_array($js_vars) && !empty($js_vars))
    <script type="text/javascript">
        window.backend_vars = {!! json_encode($js_vars) !!};
    </script>
    @endif

    @section('inline_scripts')
    @show