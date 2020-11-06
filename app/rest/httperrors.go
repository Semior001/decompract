package rest

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"

	"github.com/go-chi/render"
	log "github.com/go-pkgz/lgr"
)

// errTmplData store data for error message
type errTmplData struct {
	Error   string
	Details string
}

const errHTMLTmpl = `<!DOCTYPE html>
<html>
<head>
		<meta name="viewport" content="width=device-width"/>
		<meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
</head>
<body>
<div style="text-align: center; font-family: Arial, sans-serif; font-size: 18px;">
    <h1 style="position: relative; color: #4fbbd6; margin-top: 0.2em;">DEComPract</h1>
	<p style="position: relative; max-width: 20em; margin: 0 auto 1em auto; line-height: 1.4em;">{{.Error}}: {{.Details}}.</p>
</div>
</body>
</html>`

// SendErrorHTML makes html body with provided template and responds with provided http status code,
// error code is not included in render as it is intended for UI developers and not for the users
func SendErrorHTML(w http.ResponseWriter, r *http.Request, httpStatusCode int, err error, details string) {
	// MustExecute behaves like template.Execute, but panics if an error occurs.
	MustExecute := func(tmpl *template.Template, wr io.Writer, data interface{}) {
		if err = tmpl.Execute(wr, data); err != nil {
			panic(err)
		}
	}
	tmpl := template.Must(template.New("error").Parse(errHTMLTmpl))
	log.Printf("[WARN] %s", errDetailsMsg(r, httpStatusCode, err, details))
	render.Status(r, httpStatusCode)
	msg := bytes.Buffer{}
	MustExecute(tmpl, &msg, errTmplData{
		Error:   err.Error(),
		Details: details,
	})
	render.HTML(w, r, msg.String())
}

func errDetailsMsg(r *http.Request, code int, err error, msg string) string {
	q := r.URL.String()
	if qun, e := url.QueryUnescape(q); e == nil {
		q = qun
	}

	srcFileInfo := ""
	if pc, file, line, ok := runtime.Caller(2); ok {
		fnameElems := strings.Split(file, "/")
		funcNameElems := strings.Split(runtime.FuncForPC(pc).Name(), "/")
		srcFileInfo = fmt.Sprintf(" [caused by %s:%d %s]", strings.Join(fnameElems[len(fnameElems)-3:], "/"),
			line, funcNameElems[len(funcNameElems)-1])
	}

	remoteIP := r.RemoteAddr
	if pos := strings.Index(remoteIP, ":"); pos >= 0 {
		remoteIP = remoteIP[:pos]
	}
	if err == nil {
		err = errors.New("no error")
	}
	return fmt.Sprintf("%s - %v - %d - %s - %s%s", msg, err, code, remoteIP, q, srcFileInfo)
}
