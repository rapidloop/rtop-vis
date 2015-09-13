/*

rtop-vis - ad hoc cluster monitoring over SSH

Copyright (c) 2015 RapidLoop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"html/template"
	"log"
	"net/http"
)

var tmpl *template.Template

func startWeb() {
	tmpl = template.Must(template.New(".").Parse(html))
	http.HandleFunc("/", webServer)
	log.Fatal(http.ListenAndServe(DEFAULT_WEB_ADDR, nil))
}

func webServer(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, allStats); err != nil {
		log.Print(err)
	}
}

const html = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>rtop-viz</title>
    <link href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css" rel="stylesheet">
	<link href='https://fonts.googleapis.com/css?family=Source+Sans+Pro' rel='stylesheet' type='text/css'>
	{{$all := .}}
	<style type="text/css">
	body { background: #E8F5E9; font-family: "Source Sans Pro", sans-serif; font-size: 12px; }
	.chart {
		width: 324px; height: 200px; border-radius: 3px; background-color: #fff;
		box-shadow: 0 1px 3px rgba(0,0,0,0.12), 0 1px 2px rgba(0,0,0,0.24);
		margin: 5px;
	}
	.chartc {
		display: flex; display: -webkit-flex; flex-wrap: wrap; -webkit-flex-wrap: wrap;
	}
	.dygraph-legend { font-size: 12px !important; }
	.dygraph-title { font-size: 14px; font-weight: 400; }
	h2 { text-align: center; font-size: 24px; padding: 1.1em; }
	.footer { margin: 5em 0 2em 0; color: #aaa; text-align: center; font-size: 14px }
	a { color: #777; }
	</style>
  </head>
  <body>
  	<div class="container-fluid">
	  <div class="row">
	    <div class="col-sm-6">
			<h2>Load Average</h2>
		</div>
	    <div class="col-sm-6">
			<h2>Memory Usage</h2>
		</div>
	  </div>
	  <div class="row">
	    <div class="col-sm-6 chartc">
		  {{range .Keys}}
		  <div id="id-{{.}}" class="chart"></div>
		  {{end}}
		</div>
	    <div class="col-sm-6 chartc">
		  {{range .Keys}}
		  <div id="mid-{{.}}" class="chart"></div>
		  {{end}}
		</div>
	  </div>
	  <div class="row footer">
		<a href="https://github.com/rapidloop/rtop-viz">rtop-viz</a> is MIT-licensed open source software from <a href="https://www.rapidloop.com/">RapidLoop</a>.
		If you like this, you might like <a href="https://www.opsdash.com">OpsDash</a> - our server and service monitoring product.
	  </div>
	</div>

    <script src="https://code.jquery.com/jquery-2.1.4.min.js"></script>
    <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/js/bootstrap.min.js"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/dygraph/1.1.1/dygraph-combined.js"></script>
	<script type="text/javascript">
	$(function() {
	{{range .Keys}}{{$sr := ($all.GetRing .)}}
		new Dygraph(
			document.getElementById("id-{{.}}"),
			[
				{{range $sr.Entries}}
				[ new Date( {{.At.Unix}} * 1000 ), {{.Load1}} ],
				{{end}}
			],
			{
				title: "{{.}}",
				axisLabelFontSize: 10,
				axes: { y: { axisLabelWidth: 30 } },
				labels: [ "X", "Load Average" ],
          		gridLineColor: 'rgb(200,200,200)',
			});
		new Dygraph(
			document.getElementById("mid-{{.}}"),
			[
				{{range $sr.Entries}}
				[ new Date( {{.At.Unix}} * 1000 ),
					{{.MemCached}}, {{.MemBuffers}},
					{{.MemFree}}, {{.MemUsed}}
				],
				{{end}}
			],
			{
				title: "{{.}}",
				axisLabelFontSize: 10,
				axes: { y: { axisLabelWidth: 30 } },
				labels: [ "X", "cached", "buffers", "free", "used" ],
          		labelsKMG2: true,
				stackedGraph: true,
          		gridLineColor: 'rgb(200,200,200)',
			});
	{{end}}
	  var search = location.search || '';
	  if (search.indexOf('noreload') === -1) {
	    window.setTimeout(function() {
	      window.location.reload();
	    }, 10000);
	  }
	});
	</script>
  </body>
</html>
`
