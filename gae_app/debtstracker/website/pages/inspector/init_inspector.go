package inspector

import "github.com/julienschmidt/httprouter"

func InitInspector(router *httprouter.Router) {
	router.GET("/inspector/user", userPage)
	router.GET("/inspector/contact", contactPage{}.contactPageHandler)
	router.GET("/inspector/transfers", transfersPage{}.transfersPageHandler)
}
