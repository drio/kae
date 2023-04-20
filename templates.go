package main

var homeTmpl = `<!DOCTYPE html>
<html>
 <head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Keep an eye (home)</title>
 </head>
 <body>
  <h1>Keep an Eye</h1>

{{ if .SayHi }}
  <p>I have to say hi</p>
{{ end }}

<form method="POST" action="/newtoken" enctype="application/x-www-form-urlencoded">
 <input type="text" name="name" placeholder="name" autofocus>
 <input type="text" name="interval" placeholder="interval (secs)">
 <button>New Token</button>
</form>

<h3>List of tokens</h3>
<ul style="list-style: none;">
  {{ range .Tokens }}
  <li style="margin: 0.7em 0; {{if .Disabled}} color: silver{{end}}">
  <span style="font-weight: 800">{{ .Name }}</span> 
   {{.ID}} 
   {{.Interval}}  
   <a href="">delete</a>
   {{if .Fired}}🔥 {{end}}
   {{if .Disabled}}<a href="">enable</a> {{end}}
  </li>
  {{ end }}
</ul>

 </body>
</html>
`
