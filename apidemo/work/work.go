package work

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Work struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Tags        pq.StringArray `gorm:"type:tag[]"`
	gorm.Model
}

// set table name
func (Work) TableName() string {
	return "works"
}

// work handler
type WorkHandler struct {
	db *gorm.DB
}

func NewWorkHandler(db *gorm.DB) *WorkHandler {
	return &WorkHandler{db: db}
}

func (w *WorkHandler) NewTask(c *gin.Context) {
	var work Work
	if err := c.ShouldBindJSON(&work); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
		return
	}

	r := w.db.Create(&work)
	if err := r.Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"msg": "create",
	})
}

func (w *WorkHandler) ListWork(c *gin.Context) {
	var works []Work
	result := w.db.Find(&works)
	if err := result.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, works)

}
