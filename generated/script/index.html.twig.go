// Code generated by stickgen.
// DO NOT EDIT!

package script

import (
	"fmt"
	"io"

	"github.com/tyler-sommer/stick"
)

func blockScriptIndexHtmlTwigContent(env *stick.Env, output io.Writer, ctx map[string]stick.Value) {
	// line 3, offset 19 in script/index.html.twig
	fmt.Fprint(output, `
<h4>Scripts</h4>
<table class="table table-bordered table-striped">
	<tr>
		<th>Title</th>
		<th>Body</th>
		<th><a href="/script/new" class="btn btn-primary btn-sm">New Script</a></th>
	</tr>
	`)
	// line 11, offset 4 in script/index.html.twig
	stick.Iterate(ctx["scripts"], func(_, el stick.Value, loop stick.Loop) (brk bool, err error) {
		// line 11, offset 24 in script/index.html.twig
		fmt.Fprint(output, `
	<tr>
		<td>`)
		// line 13, offset 6 in script/index.html.twig
		{
			val, err := stick.GetAttr(el, "Title")
			if err == nil {
				fmt.Fprint(output, val)
			}
		}
		// line 13, offset 20 in script/index.html.twig
		fmt.Fprint(output, `</td>
		<td class="code-preview">`)
		// line 14, offset 27 in script/index.html.twig
		{
			val, err := stick.GetAttr(el, "Body")
			if err == nil {
				fmt.Fprint(output, val)
			}
		}
		// line 14, offset 40 in script/index.html.twig
		fmt.Fprint(output, `</td>
		<td>
			<div class="btn-group">
				<a href="/script/`)
		// line 17, offset 21 in script/index.html.twig
		{
			val, err := stick.GetAttr(el, "ID")
			if err == nil {
				fmt.Fprint(output, val)
			}
		}
		// line 17, offset 32 in script/index.html.twig
		fmt.Fprint(output, `/edit" class="btn btn-sm btn-default">Edit</a>
				<a href="/script/`)
		// line 18, offset 21 in script/index.html.twig
		{
			val, err := stick.GetAttr(el, "ID")
			if err == nil {
				fmt.Fprint(output, val)
			}
		}
		// line 18, offset 32 in script/index.html.twig
		fmt.Fprint(output, `/remove" class="remove btn btn-sm btn-warning">Remove</a>
			</div>
			`)
		// line 24, offset 14 in script/index.html.twig
		fmt.Fprint(output, `
		</td>
	</tr>
	`)
		return false, nil
	})
	// line 27, offset 13 in script/index.html.twig
	fmt.Fprint(output, `
</table>
`)
}
func blockScriptIndexHtmlTwigAdditionalJavascripts(env *stick.Env, output io.Writer, ctx map[string]stick.Value) {
	// line 31, offset 34 in script/index.html.twig
	fmt.Fprint(output, `
<script type="text/javascript">
$(function() {
	$('.remove').on('click', function(e) {
		e.preventDefault();
		
		if (confirm('Are you sure you want to delete this script?')) {
			var url = $(this).attr('href');
			$.ajax({
				url: url,
				type: 'post',
				success: function() {
					window.location.reload();
				}
			});
		}
	});

	$('.toggle').on('click', function(e) {
		e.preventDefault();

		var url = $(this).attr('href');
		$.ajax({
			url: url,
			type: 'post',
			success: function() {
				window.location.reload();
			}
		});
	});
});
</script>
`)
}

func TemplateScriptIndexHtmlTwig(env *stick.Env, output io.Writer, ctx map[string]stick.Value) {
	// line 1, offset 0 in layout.html.twig
	fmt.Fprint(output, `<!DOCTYPE html>
<html>
<head>
  <title>squIRCy</title>
  <script src="//cdn.jsdelivr.net/jquery/2.1.1/jquery.min.js"></script>
  <script src="//cdn.jsdelivr.net/bootstrap/3.2.0/js/bootstrap.min.js"></script>
  <script src="//cdn.jsdelivr.net/momentjs/2.8.1/moment.min.js"></script>
  <link rel="stylesheet" href="//cdn.jsdelivr.net/bootstrap/3.2.0/css/bootstrap.min.css">
  <link rel="stylesheet" href="//cdn.jsdelivr.net/fontawesome/4.2.0/css/font-awesome.min.css">
  <link rel="stylesheet" href="/css/style.css">
</head>

<body>
	<div id="main-container" class="container">
		<div class="row">
			<div class="col-md-12">
			`)
	// line 17, offset 6 in layout.html.twig
	blockScriptIndexHtmlTwigContent(env, output, ctx)
	// line 18, offset 17 in layout.html.twig
	fmt.Fprint(output, `
			</div>
		</div>
	</div>

	<nav id="main-nav" class="navbar navbar-default navbar-fixed-top" role="navigation">
	  	<div class="container">
			<div class="navbar-header">
				<a class="navbar-brand" href="https://github.com/tyler-sommer/squircy2">squIRCy2</a>
        	</div>
			<ul class="nav navbar-nav">
				<li><a href="/">Dashboard</a></li>
				<li><a href="/script">Scripts</a></li>
				<li><a href="/webhook">Webhooks</a></li>
				<li><a href="/manage">Settings</a></li>
				<li class="divider">&nbsp;</li>
				<li><a href="/repl">REPL</a></li>
			</ul>
			<ul class="nav navbar-nav pull-right">
				<li><a id="reinit" title="Re-initialize scripts" href="/script/reinit"><i class="fa fa-refresh"></i></a></li>
		        <li><a class="post-action" id="connect-button" style="display: none" href="/connect"><i class="fa fa-power-off"></i> Connect</a></li>
				<li><a class="post-action" id="disconnect-button" style="display: none" href="/disconnect"><i class="fa fa-power-off"></i> Disconnect</a></li>
				<li><a class="post-action" id="connecting-button" style="display: none" href="/disconnect"><i class="fa fa-power-off"></i> Connecting...</a></li>
		      </ul>
	  	</div>
	</nav>

    `)
	// line 1, offset 0 in _page_javascripts.html.twig
	fmt.Fprint(output, `<script type="text/javascript">
$(function() {
    var es = window.squIRCyEvents = new EventSource('/event');

	var $connectBtn = $('#connect-button');
	var $disconnectBtn = $('#disconnect-button');
	var $connectingBtn = $('#connecting-button');
    es.addEventListener("irc.CONNECTING", function() {
        $connectingBtn.show();
		$disconnectBtn.add($connectBtn).hide();
    });
    es.addEventListener("irc.CONNECT", function() {
        $disconnectBtn.show();
		$connectBtn.add($connectingBtn).hide();
    });
    es.addEventListener("irc.DISCONNECT", function() {
        $connectBtn.show();
        $disconnectBtn.add($connectingBtn).hide();
    });

	var $postLinks = $('.post-action');
	$postLinks.on('click', function(e) {
		e.preventDefault();
		
		var url = $(this).attr('href');
		$.ajax({
			url: url,
			type: 'post'
		});
	});
	
	var $reinit = $('#reinit').tooltip({
		placement: 'bottom',
		container: 'body'
	});
	$reinit.on('click', function(e) {
		e.preventDefault();
		
		if (confirm('Are you sure you want to reinitialize all scripts?')) {
			var url = $(this).attr('href');
			$.ajax({
				url: url,
				type: 'post'
			})
		}
	});

	$.ajax({
		url: '/status',
		type: 'get',
		dataType: 'json',
		success: function(response) {
			if (response.Status == 2) {
				$disconnectBtn.show();
				$connectBtn.add($connectingBtn).hide();
			} else if (response.Status == 1) {
				$connectingBtn.show();
				$disconnectBtn.add($connectBtn).hide();
			} else {
				$connectBtn.show();
				$disconnectBtn.add($connectingBtn).hide();
			}
		}
	});
});
</script>`)
	// line 45, offset 47 in layout.html.twig
	fmt.Fprint(output, `
    `)
	// line 46, offset 7 in layout.html.twig
	blockScriptIndexHtmlTwigAdditionalJavascripts(env, output, ctx)
	// line 47, offset 18 in layout.html.twig
	fmt.Fprint(output, `
</body>

</html>
`)
	// line 1, offset 32 in script/index.html.twig
	fmt.Fprint(output, `

`)
	// line 29, offset 14 in script/index.html.twig
	fmt.Fprint(output, `

`)
	// line 63, offset 14 in script/index.html.twig
	fmt.Fprint(output, `
`)
}
