package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"math"
	"strconv"
)

// GameBox is a gorm model for database table `gameboxes`.
type GameBox struct {
	gorm.Model
	ChallengeID uint
	TeamID      uint

	IP          string
	Port        string
	Description string
	Visible     bool
	Score       float64 // The score can be negative.
	IsDown      bool
	IsAttacked  bool
}

// GetSelfGameBoxes returns the gameboxes which belong to the team.
func (s *Service) GetSelfGameBoxes(c *gin.Context) (int, interface{}) {
	if s.Timer.Status == "wait" {
		return s.makeSuccessJSON([]int{})
	}

	var gameBoxes []struct {
		ChallengeID uint
		Title       string
		IP          string
		Port        string
		Description string
		Score       float64
		IsDown      bool
		IsAttacked  bool
	}
	teamID, _ := c.Get("teamID")

	s.Mysql.Table("game_boxes").Where(&GameBox{TeamID: teamID.(uint), Visible: true}).Order("challenge_id").Find(&gameBoxes)
	for index, gameBox := range gameBoxes {
		var challenge Challenge
		s.Mysql.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: gameBox.ChallengeID}}).Find(&challenge)
		gameBoxes[index].Title = challenge.Title
	}
	return s.makeSuccessJSON(gameBoxes)
}

// GetGameBoxes returns the gameboxes for manager.
func (s *Service) GetGameBoxes(c *gin.Context) (int, interface{}) {
	pageStr := c.Query("page")   // 当前页
	perPageStr := c.Query("per") // 每页数量

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		return s.makeErrJSON(400, 40002, "Error Query")
	}
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage <= 0 {
		return s.makeErrJSON(400, 40002, "Error Query")
	}

	var total int
	s.Mysql.Model(&GameBox{}).Count(&total)
	var gameBox []GameBox
	s.Mysql.Model(&GameBox{}).Offset((page - 1) * perPage).Limit(perPage).Find(&gameBox)

	return s.makeSuccessJSON(gin.H{
		"Data":      gameBox,
		"Total":     total,
		"TotalPage": math.Ceil(float64(total / perPage)),
	})
}

// NewGameBoxes is add a new gamebox handler for manager.
func (s *Service) NewGameBoxes(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ChallengeID uint   `binding:"required"`
		TeamID      uint   `binding:"required"`
		IP          string `binding:"required"`
		Port        string `binding:"required"`
		Description string `binding:"required"`
	}
	var inputForm []InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	for _, item := range inputForm {
		var count int

		// Check the ChallengeID
		s.Mysql.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: item.ChallengeID}}).Count(&count)
		if count != 1 {
			return s.makeErrJSON(400, 40001, "Challenge 不存在")
		}

		// Check the TeamID
		s.Mysql.Model(&Team{}).Where(&Team{Model: gorm.Model{ID: item.TeamID}}).Count(&count)
		if count != 1 {
			return s.makeErrJSON(400, 40001, "Team 不存在")
		}

		// Check if the gamebox is existed by challenge ID and team ID,
		// since every team should have only one gamebox for each challenge.
		s.Mysql.Model(GameBox{}).Where(&GameBox{ChallengeID: item.ChallengeID, TeamID: item.TeamID}).Count(&count)
		if count != 0 {
			return s.makeErrJSON(400, 40001, "存在重复添加数据")
		}
	}

	tx := s.Mysql.Begin()
	for _, item := range inputForm {
		newGameBox := &GameBox{
			ChallengeID: item.ChallengeID,
			TeamID:      item.TeamID,
			IP:          item.IP,
			Port:        item.Port,
			Description: item.Description,
		}
		if tx.Create(newGameBox).RowsAffected != 1 {
			tx.Rollback()
			return s.makeErrJSON(500, 50000, "添加 GameBox 失败！")
		}
	}
	tx.Commit()

	s.NewLog(NORMAL, "manager_operate", fmt.Sprintf("共 %d 个 GameBox 被创建", len(inputForm)))
	return s.makeSuccessJSON("添加 GameBox 成功！")
}

// EditGameBox is edit gamebox handler for manager.
func (s *Service) EditGameBox(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID uint `binding:"required"`

		IP          string `binding:"required"`
		Port        string `binding:"required"`
		Description string `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	tx := s.Mysql.Begin()
	if tx.Model(&GameBox{}).Where(&GameBox{Model: gorm.Model{ID: inputForm.ID}}).Updates(&GameBox{
		IP:          inputForm.IP,
		Port:        inputForm.Port,
		Description: inputForm.Description,
	}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50001, "修改 GameBox 失败！")
	}
	tx.Commit()

	return s.makeSuccessJSON("修改 GameBox 成功！")
}
