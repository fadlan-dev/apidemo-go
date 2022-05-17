package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/fadlan-dev/auth"
	"github.com/fadlan-dev/todo"
)

var (
	buildcommit = "dev"
	buildtime   = time.Now().String()
)

func main() {

	// Liveness Probe
	_, err := os.Create("/tmp/live")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove("/tmp/live")

	// Configuration

	if err := godotenv.Load("local.env"); err != nil && !os.IsNotExist(err) {
		log.Printf("please consider environment variables: %s\n", err)
	}

	db, err := gorm.Open(mysql.Open(os.Getenv("DB_CONN")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// auto create table
	db.AutoMigrate(&todo.Todo{})

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:8080",
	}
	config.AllowHeaders = []string{
		"Origin",
		"Authorization",
		"TransactionID",
	}

	r.Use(cors.New(config))

	// Readines Probe
	r.GET("/healthz", func(c *gin.Context) {
		c.Status(200)
	})

	// Rate Limit
	r.GET("/limitz", limitedHandler)

	// ldflags chek
	r.GET("/x", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"buildcommit": buildcommit,
			"buildtime":   buildtime,
		})
	})

	r.GET("/tokenz", auth.AccessToken(os.Getenv("SIGN")))
	protected := r.Group("", auth.Protect([]byte(os.Getenv("SIGN"))))

	// new todo handler
	t := todo.NewTodoHandler(db)
	protected.POST("/todos", t.NewTask)
	protected.GET("/todos", t.ListTodo)
	protected.GET("/todo/:id", t.GetTodo)
	protected.DELETE("/todos/:id", t.Remove)

	// new work handler
	// w := work.NewWorkHandler(db)
	// r.POST("/works", w.NewTask)
	// r.GET("/works", w.ListWork)

	// create gracefully shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	s := &http.Server{
		Addr:           ":" + os.Getenv("PORT"),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen %s\n", err)
		}
	}()

	<-ctx.Done()
	stop()
	fmt.Println("shuting down gracefully, press Ctrl + C again to force")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}
}

// Rate Limit
var limiter = rate.NewLimiter(5, 5)

func limitedHandler(c *gin.Context) {
	if !limiter.Allow() {
		c.AbortWithStatus(http.StatusTooManyRequests)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "pong",
	})
}
