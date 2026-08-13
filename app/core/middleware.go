package core

import "github.com/gin-gonic/gin"

type Middleware func(*gin.Context) error
