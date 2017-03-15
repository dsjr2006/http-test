package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	gin "gopkg.in/gin-gonic/gin.v1"
)

var db = startBolt()

// TODO: Check All Requests for Size
// TODO: Request Timeout
func startBolt() bolt.DB {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	return *db
}
func init() {
	// Init
}
func main() {
	// Debug or No
	gin.SetMode(gin.ReleaseMode)
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
		jsonG.GET("/boltdb", dbTestDisplay)
		jsonG.POST("/md5", jsonMD5)
	}
	upload := r.Group("/upload")
	{
		upload.POST("/file", uploadFile)
		upload.POST("/boltdb", dbTest)
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
	clientIP := c.ClientIP
	if clientIP == nil {
		log.Fatal("Could not obtain client IP")
		c.String(500, "Could not retrieve client IP")
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
func dbTest(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	/*
		db, err := bolt.Open("my.db", 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	*/
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("dbTest"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		timestamp := fmt.Sprintf("%v", time.Now())
		b.Put([]byte(timestamp), body)
		return nil
	})
	if err != nil {
		c.String(500, "Internal Error: Could not insert record")
		log.Fatal(err)
	}
	c.String(200, "Insert Successful")
}
func dbTestDisplay(c *gin.Context) {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("dbTest"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})
	if err != nil {
		log.Fatal("Could not display db contents")
		c.String(500, "Error: Could not display DB")
	}
	c.String(200, "Display Complete")
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
