package main

import (
	"github.com/gin-gonic/gin"
)

func (s *Service) initRouter() {
	r := gin.Default()

	// 管理员登录
	r.POST("/manager/login", func(c *gin.Context) {
		c.JSON(s.ManagerLogin(c))
	})

	// 管理
	authorized := r.Group("/manager")
	authorized.Use(s.AuthRequired())
	{
		// Challenge
		authorized.GET("/challenges", func(c *gin.Context) {
			c.JSON(s.GetAllChallenges())
		})
		authorized.POST("/challenge", func(c *gin.Context) {
			c.JSON(s.NewChallenge(c))
		})
		authorized.PUT("/challenge", func(c *gin.Context) {
			c.JSON(s.EditChallenge(c))
		})
		authorized.DELETE("/challenge", func(c *gin.Context) {
			c.JSON(s.DeleteChallenge(c))
		})
	}

	s.Router = r
	panic(r.Run(s.Conf.Base.Port))
}

// 鉴权中间件
func (s *Service) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(s.makeErrJSON(403, 40300, "未授权访问"))
			c.Abort()
			return
		}

		var managerData Manager
		s.Mysql.Where(&Manager{Token: token}).Find(&managerData)
		if managerData.ID == 0{
			c.JSON(s.makeErrJSON(401, 40100, "未授权访问"))
			c.Abort()
			return
		}

		c.Set("managerData", managerData)
		c.Next()
	}
}