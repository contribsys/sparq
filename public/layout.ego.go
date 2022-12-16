// Generated by ego.
// DO NOT EDIT

//line layout.ego:1

package public

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
	_, _ = io.WriteString(w, "\">\n  <head>\n    <meta charset=\"utf8\" />\n    <meta name=\"viewport\" content=\"width=device-width,initial-scale=1\" />\n    <title>Sparq</title>\n    <link href=\"https://cdn.jsdelivr.net/npm/bootstrap@5.2.3/dist/css/bootstrap.min.css\" rel=\"stylesheet\" integrity=\"sha384-rbsA2VBKQhggwzxH7pPCaAqO46MgnOM80zW1RWuH61DGLwZJEdK2Kadq2F9CUG65\" crossorigin=\"anonymous\">\n    <link rel=\"icon\" href=\"data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%2210 0 100 100%22><text y=%22.90em%22 font-size=%2290%22>⚡</text></svg>\">\n\n    <link href=\"/static/app.css\" media=\"screen\" rel=\"stylesheet\" type=\"text/css\" />\n    <script type=\"text/javascript\" src=\"/static/app.js\"></script>\n    <meta name=\"google\" content=\"notranslate\" />\n  </head>\n  <body data-locale=\"")
//line layout.ego:20
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(ctx(req).locale)))
//line layout.ego:20
	_, _ = io.WriteString(w, "\">\n    ")
//line layout.ego:21
	ego_nav(w, req)
//line layout.ego:22
	_, _ = io.WriteString(w, "\n    \n    ")
//line layout.ego:23
	yield()
//line layout.ego:24
	_, _ = io.WriteString(w, "\n    \n    ")
//line layout.ego:25
	ego_footer(w, req)
//line layout.ego:26
	_, _ = io.WriteString(w, "\n    <script src=\"https://cdn.jsdelivr.net/npm/bootstrap@5.2.3/dist/js/bootstrap.bundle.min.js\" integrity=\"sha384-kenU1KFdBIe4zVF0s0G1M5b4hcpxyD9F7jL+jjXkk+Q2h455rYXK/7HAuoJl+0I4\" crossorigin=\"anonymous\"></script>\n  </body>\n</html>\n")
//line layout.ego:29
}

var _ fmt.Stringer
var _ io.Reader
var _ context.Context
var _ = html.EscapeString