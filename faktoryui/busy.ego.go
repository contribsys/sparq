// Generated by ego.
// DO NOT EDIT

//line busy.ego:1

package faktoryui

import "fmt"
import "html"
import "io"
import "context"

import (
	"net/http"

	"github.com/contribsys/faktory/manager"
)

func ego_busy(w io.Writer, req *http.Request) {

//line busy.ego:12
	_, _ = io.WriteString(w, "\n\n")
//line busy.ego:13
	ego_layout(w, req, func() {
//line busy.ego:14
		_, _ = io.WriteString(w, "\n\n<div class=\"row header mt-3\">\n  <div class=\"col-12\">\n    <h3>")
//line busy.ego:17
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Jobs"))))
//line busy.ego:17
		_, _ = io.WriteString(w, "</h3>\n  </div>\n</div>\n\n<div class=\"table-responsive\">\n  <table class=\"workers table table-hover table-bordered table-striped table-light\">\n    <thead>\n      <th>")
//line busy.ego:24
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Process"))))
//line busy.ego:24
		_, _ = io.WriteString(w, "</th>\n      <th>")
//line busy.ego:25
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "JID"))))
//line busy.ego:25
		_, _ = io.WriteString(w, "</th>\n      <th>")
//line busy.ego:26
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Queue"))))
//line busy.ego:26
		_, _ = io.WriteString(w, "</th>\n      <th>")
//line busy.ego:27
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Job"))))
//line busy.ego:27
		_, _ = io.WriteString(w, "</th>\n      <th>")
//line busy.ego:28
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Arguments"))))
//line busy.ego:28
		_, _ = io.WriteString(w, "</th>\n      <th>")
//line busy.ego:29
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Started"))))
//line busy.ego:29
		_, _ = io.WriteString(w, "</th>\n    </thead>\n    ")
//line busy.ego:31
		busyReservations(req, func(res *manager.Reservation) {
//line busy.ego:32
			_, _ = io.WriteString(w, "\n      ")
//line busy.ego:32
			job := res.Job
//line busy.ego:33
			_, _ = io.WriteString(w, "\n      <tr>\n        <td>\n          <code>\n            ")
//line busy.ego:36
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(res.Wid)))
//line busy.ego:37
			_, _ = io.WriteString(w, "\n          </code>\n        </td>\n        <td>\n          <code>\n            ")
//line busy.ego:41
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(job.Jid)))
//line busy.ego:42
			_, _ = io.WriteString(w, "\n          </code>\n        </td>\n        <td>\n          <a href=\"")
//line busy.ego:45
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(root(req))))
//line busy.ego:45
			_, _ = io.WriteString(w, "/queues/")
//line busy.ego:45
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(job.Queue)))
//line busy.ego:45
			_, _ = io.WriteString(w, "\">")
//line busy.ego:45
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(job.Queue)))
//line busy.ego:45
			_, _ = io.WriteString(w, "</a>\n        </td>\n        <td><code>")
//line busy.ego:47
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(job.Type)))
//line busy.ego:47
			_, _ = io.WriteString(w, "</code></td>\n        <td>\n          <div class=\"args\">")
//line busy.ego:49
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(displayArgs(job.Args))))
//line busy.ego:49
			_, _ = io.WriteString(w, "</div>\n        </td>\n        <td>")
//line busy.ego:51
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(relativeTime(res.Since))))
//line busy.ego:51
			_, _ = io.WriteString(w, "</td>\n      </tr>\n    ")
//line busy.ego:53
		})
//line busy.ego:54
		_, _ = io.WriteString(w, "\n  </table>\n</div>\n")
//line busy.ego:56
	})
//line busy.ego:57
	_, _ = io.WriteString(w, "\n")
//line busy.ego:57
}

var _ fmt.Stringer
var _ io.Reader
var _ context.Context
var _ = html.EscapeString