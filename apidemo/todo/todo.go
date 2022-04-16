package todo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Todo struct {
	Title string `json:"text"`
	gorm.Model
}

// table name
func (Todo) TableName() string {
	return "todos"
}

// todo handler
type TodoHandler struct {
	db *gorm.DB
}

// new todo handler
func NewTodoHandler(db *gorm.DB) *TodoHandler {
	return &TodoHandler{db: db}
}

func (t *TodoHandler) NewTask(c *gin.Context) {
	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
		return
	}

	r := t.db.Create(&todo)
	if err := r.Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"ID": todo.Model.ID,
	})
}

func (w *TodoHandler) ListTodo(c *gin.Context) {
	var todos []Todo
	result := w.db.Find(&todos)
	if err := result.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, todos)

}
