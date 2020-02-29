package webserver

func (ws *WebServer) muxRouter() {
	// ws.router.GET("/", ws.Recovery(ws.helloWorldGetHandler()))
	// ws.router.GET("/get", ws.Recovery(ws.simpleGetHandler()))
	ws.router.POST("/test", ws.Recovery(ws.testHandler()))
}
