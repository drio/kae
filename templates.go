package main

var homeTmpl = `<!DOCTYPE html>
<html>
 <head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Keep an eye (home)</title>
  <link rel="icon" type="image/x-icon" href="/assets/favicon-32x32.png">
  <link rel="stylesheet" href="/assets/style.css">
  </head>
<body style="padding: 1rem">

  <h1>Keep an Eye</h1>

  {{ if .SayHi }}
    <p>I have to say hi</p>
  {{ end }}

  <form method="POST" action="/newtoken" enctype="application/x-www-form-urlencoded">
   <input type="text" name="name" placeholder="name" autofocus> <br/>
   <input type="text" name="interval" placeholder="interval (secs)"> <br/>
   <input type="text" name="description" placeholder="description" style="width: 100%; max-width: 400px"> <br/>
   <button>New Token</button>
  </form>

  <div class="tokens-container">
  {{ range .Tokens }}
  <div class="entry" style="{{if .Disabled}} color: silver{{end}}">
    <div> 
      {{if not .Disabled}}
        <span class="emoji">{{if .Fired}}🔥{{else}}🟢{{end}}</span>
      {{end}}
      <span class="token-name">{{ .Name }}</span>
    </div>
   <div class="token-value">{{ .Token }}</div>

   <div>({{.Interval}}s)</div>

   <div>{{.Description}}</div>

   <div>
    <a href="/delete/{{.ID}}" class="danger">delete</a> |
    {{if .Disabled}}
    <a href="/enable/{{.ID}}">enable</a> 
    {{else}}
    <a href="/disable/{{.ID}}">disable</a>
    {{end}}
    </div>
  </div>
  {{ end }}
</div>

 </body>
</html>
`
