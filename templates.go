package main

var homeTmpl = `<!DOCTYPE html>
<html>
 <head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Keep an eye (home)</title>
 </head>
  <body style="padding: 1rem">

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
  <ul style="list-style: none; padding-left: 1rem">
  {{ range .Tokens }}
  <li style="margin: 0.5em 0; {{if .Disabled}} color: silver{{end}}">
   {{if .Fired}}🔥 {{end}}
   <span style="font-weight: 800">{{ .Name }}</span> 
   <span style="color: silver">{{ .ID }}</span> 
   {{.Interval}}s  
   <a href="/">delete</a>
   {{if .Disabled}}<a href="">enable</a> {{end}}
  </li>
  {{ end }}
</ul>

 </body>
</html>
`
