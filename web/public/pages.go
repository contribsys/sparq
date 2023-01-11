package public

import "github.com/contribsys/sparq/web"

func init() {
	// these are the pages which can be rendered
	web.RegisterPages("public/index", "public/profile", "public/home", "public/login")
}
