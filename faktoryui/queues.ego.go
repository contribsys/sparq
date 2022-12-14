// Generated by ego.
// DO NOT EDIT

//line queues.ego:1

package faktoryui

import "fmt"
import "html"
import "io"
import "context"

import "net/http"

func ego_listQueues(w io.Writer, req *http.Request) {

//line queues.ego:8
	_, _ = io.WriteString(w, "\n\n")
//line queues.ego:9
	ego_layout(w, req, func() {
//line queues.ego:10
		_, _ = io.WriteString(w, "\n\n<h3>")
//line queues.ego:11
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Queues"))))
//line queues.ego:11
		_, _ = io.WriteString(w, "</h3>\n\n<div class=\"table-responsive\">\n  <table class=\"queues table table-hover table-bordered table-striped table-light\">\n    <thead>\n      <th>")
//line queues.ego:16
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Queue"))))
//line queues.ego:16
		_, _ = io.WriteString(w, "</th>\n      <th>")
//line queues.ego:17
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Size"))))
//line queues.ego:17
		_, _ = io.WriteString(w, "</th>\n      <th>")
//line queues.ego:18
		_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Actions"))))
//line queues.ego:18
		_, _ = io.WriteString(w, "</th>\n    </thead>\n    ")
//line queues.ego:20
		for _, queue := range queues(req) {
//line queues.ego:21
			_, _ = io.WriteString(w, "\n      <tr>\n        <td>\n          <a href=\"")
//line queues.ego:23
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(root(req))))
//line queues.ego:23
			_, _ = io.WriteString(w, "/queues/")
//line queues.ego:23
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(queue.Name)))
//line queues.ego:23
			_, _ = io.WriteString(w, "\">")
//line queues.ego:23
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(queue.Name)))
//line queues.ego:23
			_, _ = io.WriteString(w, "</a>\n        </td>\n        <td>")
//line queues.ego:25
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(uintWithDelimiter(queue.Size))))
//line queues.ego:25
			_, _ = io.WriteString(w, "</td>\n        <td class=\"delete-confirm\">\n          <form action=\"")
//line queues.ego:27
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(root(req))))
//line queues.ego:27
			_, _ = io.WriteString(w, "/queues/")
//line queues.ego:27
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(queue.Name)))
//line queues.ego:27
			_, _ = io.WriteString(w, "\" method=\"post\">\n            ")
//line queues.ego:28
			_, _ = fmt.Fprint(w, csrfTag(req))
//line queues.ego:29
			_, _ = io.WriteString(w, "\n            <button class=\"btn btn-danger btn-sm\" type=\"submit\" name=\"action\" value=\"delete\" data-confirm=\"")
//line queues.ego:29
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "AreYouSure"))))
//line queues.ego:29
			_, _ = io.WriteString(w, "\">")
//line queues.ego:29
			_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "ClearQueue"))))
//line queues.ego:29
			_, _ = io.WriteString(w, "</button>\n            ")
//line queues.ego:30
			if queue.IsPaused {
//line queues.ego:31
				_, _ = io.WriteString(w, "\n              <button class=\"btn btn-primary btn-sm\" type=\"submit\" name=\"action\" value=\"resume\">")
//line queues.ego:31
				_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Resume"))))
//line queues.ego:31
				_, _ = io.WriteString(w, "</button>\n            ")
//line queues.ego:32
			} else {
//line queues.ego:33
				_, _ = io.WriteString(w, "\n              <button class=\"btn btn-primary btn-sm\" type=\"submit\" name=\"action\" value=\"pause\">")
//line queues.ego:33
				_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(t(req, "Pause"))))
//line queues.ego:33
				_, _ = io.WriteString(w, "</button>\n            ")
//line queues.ego:34
			}
//line queues.ego:35
			_, _ = io.WriteString(w, "\n          </form>\n        </td>\n      </tr>\n    ")
//line queues.ego:38
		}
//line queues.ego:39
		_, _ = io.WriteString(w, "\n  </table>\n</div>\n\n  ")
//line queues.ego:42
	})
//line queues.ego:43
	_, _ = io.WriteString(w, "\n")
//line queues.ego:43
}

var _ fmt.Stringer
var _ io.Reader
var _ context.Context
var _ = html.EscapeString
