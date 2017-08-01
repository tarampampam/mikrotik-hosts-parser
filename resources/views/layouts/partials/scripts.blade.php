    <script type="text/javascript" src="{{ url('/components/flat-ui/dist/js/vendor/jquery.min.js') }}"></script>
    <script type="text/javascript" src="{{ url('/components/flat-ui/dist/js/flat-ui.min.js') }}"></script>

    @if(isset($js_vars) && is_array($js_vars) && !empty($js_vars))
    <script type="text/javascript">
        window.backend_vars = {!! json_encode($js_vars) !!};
    </script>
    @endif

    @section('inline_scripts')
    @show