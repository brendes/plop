plop
====

Plop is another static website generator with no configuration.

Why? Because, it's faster to build your own website generator than to learn how to use existing ones.

Features
--------

- ~200 lines of code
- Written in Go
- No configuration, no imposed directory layout
- Convert Markdown files to HTML from a template
- Uses Golang template engine
- Generate RSS feed

Installation
------------

```bash
go install
```

Example
-------

Create a file `src/homepage.md`:

```
title: Some Title
description: Some article description
template: homepage
uri: index.html
---

My article content.
```

Create a template file `templates/homepage.html`:

```html
{{define "homepage"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <title>{{ .Title }}</title>
    {{ if .Description }}
        <meta name="description" content="{{ .Description }}">
    {{ end }}

    <link rel="shortcut icon" href="/favicon.ico" type="image/x-icon">
    <link rel="stylesheet" type="text/css" href="/stylesheet.css">
</head>
<body>
    {{ .Body | markdown }}
</body>
</html>
{{ end }}
```

Run `plop` to generate the static file:

```bash
plop
```

Run a local HTTP server to preview the result with `plop -serve`.

### Header Properties

- `title`: Page title
- `description`: Optional article description
- `template`: Name of the Golang template
- `uri`: HTML file to generate
- `date`: Date (ISO-8601)
- `rss`: Boolean to indicate if the article should be included into a RSS feed

### Command Line Arguments

```bash
$ plop -h
Usage of plop:
  -port string
        Port to listen on (default "3000")
  -serve
        Start local HTTP server
  -src string
        Sources folder (default "src")
  -tpl string
        Templates folder (default "templates")
```
