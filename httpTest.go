package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"

	gin "gopkg.in/gin-gonic/gin.v1"
)

// TODO: Check All Requests for Size
// TODO: Request Timeout

func main() {
	// Router Config
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "static")
	// Routes
	r.GET("/", index)
	r.GET("/ping", ping)
	r.GET("/status:code", status)
	// Route Groups
	jsonG := r.Group("/json")
	{
		jsonG.GET("/", jsonR)
		jsonG.GET("/ip", jsonIP)
		jsonG.GET("/client", jsonClient)
		jsonG.GET("/echo", jsonEcho)
		jsonG.POST("/md5", jsonMD5)
	}
	upload := r.Group("/upload")
	{
		upload.POST("/file", uploadFile)
	}
	// Start Server
	r.Run(":8088")
}

func index(c *gin.Context) {
	c.HTML(200, "index.tmpl", gin.H{"title": "HTTP Test"})
}
func ping(c *gin.Context) {
	c.String(200, "pong")
}
func status(c *gin.Context) {
	val := c.Param("code")
	if intVal, err := strconv.Atoi(val); err == nil {
		c.Status(intVal)
	}
	c.Status(400)
}
func jsonR(c *gin.Context) {
	// Return JSON response
	c.IndentedJSON(200, gin.H{
		"Title":   "Test Data",
		"Message": "What do I have to say",
	})
}

// Returns JSON of Client IP *Does not attempt to check for proxy
func jsonIP(c *gin.Context) {
	clientIP, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		log.Fatal("Could not obtain client IP")
	}
	c.IndentedJSON(200, gin.H{
		"ip": clientIP,
	})
}

// Returns JSON of Client Info
func jsonClient(c *gin.Context) {
	// Check for UA
	if ua := c.Request.Header.Get("User-Agent"); ua != "" {
		c.IndentedJSON(200, gin.H{
			"User-Agent": ua,
		})
	} else {
		c.Header("X-HTTP-Test-Error", "No UA from Client")
		c.IndentedJSON(400, gin.H{"Error": "Could not read User-Agent from Client"})
	}
}
func jsonEcho(c *gin.Context) {
	var header []string
	for k, v := range c.Request.Header {
		log.Println("key:", k, "value:", v)
		header = append(header, fmt.Sprintf("%v:%v", k, v))
	}
	c.IndentedJSON(200, gin.H{
		"From Client": header,
	})
}
func jsonMD5(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(500)
	}
	hash := md5.Sum(body)
	c.IndentedJSON(200, gin.H{
		"MD5": hex.EncodeToString(hash[:]),
	})
}
func uploadFile(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	fmt.Printf("File Received, size: %v\n", len(body))
	if err != nil {
		c.Status(500)
	}
	c.String(200, "Upload successful")
}
