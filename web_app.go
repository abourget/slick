package main

import (
	"html/template"
)

var webappCode = template.Must(template.New("").Parse(`
<html><head><script>USER = {{.AsJavascript}}</script>
</head><body>

    Ok, folks, welcome.

</body></html>
`))
