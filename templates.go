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

  {{.Name}}

<form method="POST" enctype="application/x-www-form-urlencoded">
 <input type="text" name="name" placeholder="name" autofocus>
 <input type="text" name="interval" placeholder="interval (secs)">
 <button>New Token</button>
</form>

<h3>Tokens</h3>
{{ range .ListTokens }}
<li style="margin: 0.7em 0">
 <a href="/lists/{{ .ID }}">{{ .Name }}</a> delete
</li>
{{ end }}

 </body>
</html>
`
