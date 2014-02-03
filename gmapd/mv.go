package main

import (
	"github.com/peak6/gmap"
	. "github.com/peak6/logger"
	"html/template"
	"net/http"
)

var mvSrc = `
<html>
<head>
	<title>MapView</title>
	<link rel="stylesheet" href="//code.jquery.com/ui/1.10.4/themes/smoothness/jquery-ui.css">
	<script src="//code.jquery.com/jquery-1.9.1.js"></script>
	<script src="//code.jquery.com/ui/1.10.4/jquery-ui.js"></script>
	<script>
		$(function() {
			$( "#accordion" )
			.accordion({
				header: "> div > h3",
				collapsible: true,
				active: false

			})
			.sortable({
				axis: "y",
				handle: "h3",
				stop: function( event, ui ) {
					ui.item.children( "h3" ).triggerHandler( "focusout" ); // IE doesn't register the blur when sorting
				}
			});
		});
	</script>
</head>
<body>
<div id="accordion">
	{{ range $path, $om := .Data}}
		<div class="group">
			<h3>{{$path}} ({{len $om.Data}})</h3>
			<div>
				<table class="entry">
					<tr>
						<th>Node</th><th>Client</th><th>Value</th>
					</tr>
					{{ range $owner, $val := $om.Data}}
					<tr>
						<td>{{$owner.Node.Name}}</td>
						<td>{{$owner.Client}}</td>
						<td>{{$val}}</td>
					</tr>
					{{end}}
				</table>
			</div>
		</div>
	{{end}}
</div>
</body>
</html>
`
var mvTemplate *template.Template

func init() {
	var err error
	mvTemplate, err = template.New("mv").Parse(mvSrc)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/mv", mapView)
}

func mapView(w http.ResponseWriter, r *http.Request) {
	err := mvTemplate.Execute(w, gmap.MyStore)
	if err != nil {
		Lerr.Println("Failed to render template", err)
	}
}
