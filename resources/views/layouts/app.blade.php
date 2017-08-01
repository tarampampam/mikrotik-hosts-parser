<!DOCTYPE html>
<html lang="{{ config('app.locale', 'en') }}">
<head>
    @include('layouts.partials.html_header')
</head>
<body>

@include('layouts.partials.main_header')

<div class="container">
    @include('layouts.partials.backend_messages')
    @yield('main_content')
</div>

@if (!isset($without_footer) || $without_footer !== true)
    @include('layouts.partials.footer')
@endif
@section('scripts')
    @include('layouts.partials.scripts')
@show
</body>
</html>