// Generated by ego.
// DO NOT EDIT

//line retry.ego:1

package faktoryui

import "fmt"
import "html"
import "io"
import "context"

import (
	"net/http"

	"github.com/contribsys/faktory/client"
)

func ego_retry(w io.Writer, req *http.Request, key string, retry *client.Job) {

//line retry.ego:12
	_, _ = io.WriteString(w, "\n\n")
//line retry.ego:13
	ego_layout(w, req, func() {
//line retry.ego:14
		_, _ = io.WriteString(w, "\n\n")
//line retry.ego:15
		ego_job_info(w, req, retry)
//line retry.ego:16
		_, _ = io.WriteString(w, "\n\n<h3>")
//line retry.ego:17
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Error"))))
//line retry.ego:17
		_, _ = io.WriteString(w, "</h3>\n<div class=\"table-responsive\">\n  <table class=\"error table table-bordered table-striped table-light\">\n    <tbody>\n      <tr>\n        <th>")
//line retry.ego:22
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "ErrorClass"))))
//line retry.ego:22
		_, _ = io.WriteString(w, "</th>\n        <td>\n          <code>")
//line retry.ego:24
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(retry.Failure.ErrorType)))
//line retry.ego:24
		_, _ = io.WriteString(w, "</code>\n        </td>\n      </tr>\n      <tr>\n        <th>")
//line retry.ego:28
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "ErrorMessage"))))
//line retry.ego:28
		_, _ = io.WriteString(w, "</th>\n        <td>")
//line retry.ego:29
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(retry.Failure.ErrorMessage)))
//line retry.ego:29
		_, _ = io.WriteString(w, "</td>\n      </tr>\n      ")
//line retry.ego:31
		if retry.Failure.Backtrace != nil {
//line retry.ego:32
			_, _ = io.WriteString(w, "\n        <tr>\n          <th>")
//line retry.ego:33
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "ErrorBacktrace"))))
//line retry.ego:33
			_, _ = io.WriteString(w, "</th>\n          <td>\n            <code>\n              ")
//line retry.ego:36
			for _, line := range retry.Failure.Backtrace {
//line retry.ego:37
				_, _ = io.WriteString(w, "\n                ")
//line retry.ego:37
				_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(line)))
//line retry.ego:37
				_, _ = io.WriteString(w, "<br/>\n              ")
//line retry.ego:38
			}
//line retry.ego:39
			_, _ = io.WriteString(w, "\n            </code>\n          </td>\n        </tr>\n      ")
//line retry.ego:42
		}
//line retry.ego:43
		_, _ = io.WriteString(w, "\n    </tbody>\n  </table>\n</div>\n\n<form class=\"form-horizontal\" action=\"")
//line retry.ego:47
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(root(req))))
//line retry.ego:47
		_, _ = io.WriteString(w, "/retries/")
//line retry.ego:47
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(key)))
//line retry.ego:47
		_, _ = io.WriteString(w, "\" method=\"post\">\n  ")
//line retry.ego:48
		_, _ = fmt.Fprint(w, csrfTag(req))
//line retry.ego:49
		_, _ = io.WriteString(w, "\n  <div class=\"pull-left\">\n    <a class=\"btn btn-default\" href=\"")
//line retry.ego:50
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(root(req))))
//line retry.ego:50
		_, _ = io.WriteString(w, "/retries\">")
//line retry.ego:50
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "GoBack"))))
//line retry.ego:50
		_, _ = io.WriteString(w, "</a>\n    <button class=\"btn btn-primary btn-sm\" type=\"submit\" name=\"action\" value=\"retry\">")
//line retry.ego:51
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "RetryNow"))))
//line retry.ego:51
		_, _ = io.WriteString(w, "</button>\n    <button class=\"btn btn-danger btn-sm\" type=\"submit\" name=\"action\" value=\"delete\">")
//line retry.ego:52
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Delete"))))
//line retry.ego:52
		_, _ = io.WriteString(w, "</button>\n  </div>\n</form>\n")
//line retry.ego:55
	})
//line retry.ego:56
	_, _ = io.WriteString(w, "\n")
//line retry.ego:56
}

var _ fmt.Stringer
var _ io.Reader
var _ context.Context
var _ = html.EscapeString
