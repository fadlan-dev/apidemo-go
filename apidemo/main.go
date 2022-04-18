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

	"github.com/fadlan-dev/auth"
	"github.com/fadlan-dev/todo"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	defer os.Remove("/tpm/live")

	// Configuration
	err = godotenv.Load("local.env")
	if err != nil {
		fmt.Printf("please consider enviroment varibles: %s /n", err)
	}

	db, err := gorm.Open(mysql.Open(os.Getenv("DB_CONN")), &gorm.Config{})
	if err != nil {
		panic("failed to connet database " + err.Error())
	}

	// auto create table
	db.AutoMigrate(&todo.Todo{})

	r := gin.Default()

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
