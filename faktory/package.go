/*
This is a custom packaging of the Faktory job engine which can be embedded within
Sparq in order to provide in-process persistent background jobs.

We do not provide a TCP port for external access. All job creation and processing
happens within the Sparq process. The Faktory Web UI is mounted within the Sparq
Web UI just as Mastodon provides the Sidekiq Web UI.
*/
package faktory
