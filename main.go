package main

import (
	"./dbwrapper"
	"./routes"
	"./sessions"
	"github.com/gin-gonic/gin"
	"log"
	"os/exec"
	"runtime"
	"time"
)

var sm sessions.SessionManager

func Middle(c *gin.Context) {
	log.Println("Connected", c.ClientIP())
	id := sm.GetCookie(c.Request, c.Writer)
	err, currentSession := sm.GetSession(id)
	if err != nil {
		currentSession = sessions.CreateSession(id, time.Now().Add(24*time.Hour))
		sm.SetSession(currentSession)
	}
	c.Set("currentSession", currentSession)
	c.Set("database", &dbwrapper.Wrapper)
	c.Next()
	sm.SetSession(currentSession)
	log.Println("Disconnected", c.ClientIP())
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func main() {
	log.Println("Start listening 8080")
	err := dbwrapper.Wrapper.ReplaceRequestList("requests.sqls")
	if err != nil {
		log.Panicln("Sql error:", err)
	}
	sm.OpenSessionManager()
	defer sm.CloseSessionManager()
	r := gin.Default()
	r.Use(Middle)
	r.GET("/", routes.GetIndexHandler)
	r.GET("/registration", routes.GetRegistrationHandler)
	r.POST("/registration", routes.PostRegistrationHandler)
	r.GET("/login", routes.GetLoginHandler)
	r.POST("/login", routes.PostLoginHandler)
	r.GET("/change_password", routes.GetChangePasswordHandler)
	r.POST("/change_password", routes.PostChangePasswordHandler)
	r.GET("/config", routes.GetConfigHandler)
	r.POST("/config", routes.PostConfigHandler)
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./assets")
	go open("http://localhost:8080/")
	r.Run(":8080")
}
