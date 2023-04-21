package main

var homeTmpl = `<!DOCTYPE html>
<html>
 <head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Keep an eye (home)</title>
  <link rel="icon" type="image/x-icon" href="/assets/favicon-32x32.png">
  <style>
  body {
    font-size: 1.1rem;
    padding: 2rem;
  }
  a {
    text-decoration: none;
    color: steelblue;
  }
  .danger {
    color: red;
    font-weight: 800;
  }

  form > input {
    margin-top: .5rem;
  }

  form > button {
    margin-top: .5rem;
  }

  ul {
    list-style: none;
    padding: 0;
  }

  li {
    padding-bottom: 1rem;
  }
  </style>

 </head>
<body style="padding: 1rem">

  <h1>Keep an Eye</h1>

  {{ if .SayHi }}
    <p>I have to say hi</p>
  {{ end }}

  <form method="POST" action="/newtoken" enctype="application/x-www-form-urlencoded">
   <input type="text" name="name" placeholder="name" autofocus> <br/>
   <input type="text" name="interval" placeholder="interval (secs)"> <br/>
   <button>New Token</button>
  </form>

  <h3>List of tokens</h3>
  <ul>
  {{ range .Tokens }}
  <li style="{{if .Disabled}} color: silver{{end}}">

   {{if not .Disabled}}
    {{if .Fired}}🔥{{else}}🟢{{end}}
   {{end}}

   <span style="font-weight: 800">{{ .Name }}</span>
   <span style="color: silver">{{ .Token }}</span>

   ({{.Interval}}s)


   <br/>
   <a href="/delete/{{.ID}}" class="danger">delete</a> |
   {{if .Disabled}}
    <a href="/enable/{{.ID}}">enable</a> 
  {{else}}
    <a href="/disable/{{.ID}}">disable</a>
  {{end}}
  </li>
  {{ end }}
</ul>

 </body>
</html>
`
