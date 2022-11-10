// Generated by ego.
// DO NOT EDIT

//line layout.ego:1

package faktoryui

import "fmt"
import "html"
import "io"
import "context"
import "net/http"

func ego_layout(w io.Writer, req *http.Request, yield func()) {

//line layout.ego:7
	_, _ = io.WriteString(w, "\n<!doctype html>\n<html lang=\"")
//line layout.ego:8
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(ctx(req).locale)))
//line layout.ego:8
	_, _ = io.WriteString(w, "\" dir=\"")
//line layout.ego:8
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(textDir(req))))
//line layout.ego:8
	_, _ = io.WriteString(w, "\">\n  <head>\n    <title>")
//line layout.ego:10
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(productTitle(req))))
//line layout.ego:10
	_, _ = io.WriteString(w, "</title>\n    <meta charset=\"utf8\" />\n    <link rel=\"shortcut icon\" href=\"")
//line layout.ego:12
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(relative(req, "/static/img/favicon.ico"))))
//line layout.ego:12
	_, _ = io.WriteString(w, "\">\n    <link rel=\"mask-icon\" href=\"")
//line layout.ego:13
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(relative(req, "/static/img/favicon.svg"))))
//line layout.ego:13
	_, _ = io.WriteString(w, "\" color=\"#000000\">\n    <link rel=\"apple-touch-icon\" sizes=\"180x180\" href=\"")
//line layout.ego:14
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(relative(req, "/static/img/apple-touch-icon.png"))))
//line layout.ego:14
	_, _ = io.WriteString(w, "\">\n    <link rel=\"icon\" type=\"image/png\" sizes=\"32x32\" href=\"")
//line layout.ego:15
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(relative(req, "/static/img/favicon-32x32.png"))))
//line layout.ego:15
	_, _ = io.WriteString(w, "\">\n    <link rel=\"icon\" type=\"image/png\" sizes=\"16x16\" href=\"")
//line layout.ego:16
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(relative(req, "/static/img/favicon-16x16.png"))))
//line layout.ego:16
	_, _ = io.WriteString(w, "\">\n\n    <meta name=\"viewport\" content=\"width=device-width,initial-scale=1.0\" />\n\n    ")
//line layout.ego:20
	if rtl(req) {
//line layout.ego:21
		_, _ = io.WriteString(w, "\n    <link href=\"")
//line layout.ego:21
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(relative(req, "/static/bootstrap-rtl.min.css"))))
//line layout.ego:21
		_, _ = io.WriteString(w, "\" media=\"screen\" rel=\"stylesheet\" type=\"text/css\"/>\n    ")
//line layout.ego:22
	} else {
//line layout.ego:23
		_, _ = io.WriteString(w, "\n    <link href=\"")
//line layout.ego:23
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(relative(req, "/static/bootstrap.css"))))
//line layout.ego:23
		_, _ = io.WriteString(w, "\" media=\"screen\" rel=\"stylesheet\" type=\"text/css\" />\n    ")
//line layout.ego:24
	}
//line layout.ego:25
	_, _ = io.WriteString(w, "\n\n    <link href=\"")
//line layout.ego:26
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(relative(req, "/static/application.css"))))
//line layout.ego:26
	_, _ = io.WriteString(w, "\" media=\"screen\" rel=\"stylesheet\" type=\"text/css\" />\n    ")
//line layout.ego:27
	if rtl(req) {
//line layout.ego:28
		_, _ = io.WriteString(w, "\n    <link href=\"")
//line layout.ego:28
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(relative(req, "/static/application-rtl.css"))))
//line layout.ego:28
		_, _ = io.WriteString(w, "\" media=\"screen\" rel=\"stylesheet\" type=\"text/css\" />\n    ")
//line layout.ego:29
	}
//line layout.ego:30
	_, _ = io.WriteString(w, "\n    ")
//line layout.ego:30
	_, _ = fmt.Fprint(w, extraCss(req))
//line layout.ego:31
	_, _ = io.WriteString(w, "\n\n    <script type=\"text/javascript\" src=\"")
//line layout.ego:32
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(relative(req, "/static/application.js"))))
//line layout.ego:32
	_, _ = io.WriteString(w, "\"></script>\n    <meta name=\"google\" content=\"notranslate\" />\n  </head>\n  <body class=\"admin d-flex flex-column\" data-locale=\"")
//line layout.ego:35
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(ctx(req).locale)))
//line layout.ego:35
	_, _ = io.WriteString(w, "\" data-poll-path=\"\">\n    ")
//line layout.ego:36
	ego_nav(w, req)
//line layout.ego:37
	_, _ = io.WriteString(w, "\n    <div id=\"page\" class=\"flex-fill\">\n      <div class=\"container-xl\">\n        <div class=\"row\">\n          <div class=\"col-12 summary_bar\">\n            ")
//line layout.ego:41
	ego_summary(w, req)
//line layout.ego:42
	_, _ = io.WriteString(w, "\n          </div>\n\n          <div class=\"col-12\">\n            ")
//line layout.ego:45
	yield()
//line layout.ego:46
	_, _ = io.WriteString(w, "\n          </div>\n        </div>\n      </div>\n    </div>\n    ")
//line layout.ego:50
	ego_footer(w, req)
//line layout.ego:51
	_, _ = io.WriteString(w, "\n  </body>\n</html>\n")
//line layout.ego:53
}

var _ fmt.Stringer
var _ io.Reader
var _ context.Context
var _ = html.EscapeString
