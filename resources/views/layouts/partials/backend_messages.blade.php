@if (isset($errors) && count($errors) > 0)
<div class="alert alert-danger alert-dismissible">
  <button type="button" class="close" data-dismiss="alert" aria-hidden="true">&#x2715;</button>
  <h4><i class="icon fa fa-ban"></i> Errors!</h4>
  Dude, we have some problems:
  <ul>
    @foreach ($errors->all() as $error)
    <li>{{ $error }}</li>
    @endforeach
  </ul>
</div>
@endif
