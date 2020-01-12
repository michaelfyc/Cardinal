package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/thanhpk/randstr"
)

type Token struct {
	gorm.Model

	TeamID uint
	Token  string
}

type Team struct {
	gorm.Model

	Name      string
	Password  string `json:"-"`
	Logo      string
	Score     int64
	SecretKey string
}

func (s *Service) TeamLogin(c *gin.Context) (int, interface{}) {
	type TeamLoginForm struct {
		Name     string `json:"Name" binding:"required"`
		Password string `json:"Password" binding:"required"`
	}

	var formData TeamLoginForm
	err := c.BindJSON(&formData)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	var team Team
	s.Mysql.Where(&Team{Name: formData.Name}).Find(&team)
	if team.Name != "" && s.checkPassword(formData.Password, team.Password) {
		// 登录成功
		token := s.generateToken()

		tx := s.Mysql.Begin()
		if tx.Create(&Token{TeamID: team.ID, Token: token}).RowsAffected != 1 {
			tx.Rollback()
			return s.makeErrJSON(500, 50000, "Server error")
		}
		tx.Commit()
		return s.makeSuccessJSON(token)
	} else {
		return s.makeErrJSON(403, 40300, "账号或密码错误！")
	}
}

func (s *Service) GetTeamInfo(c *gin.Context) (int, interface{}){
	teamID, ok := c.Get("teamID")
	if !ok{
		return s.makeErrJSON(500, 50000, "Server error")
	}

	var teamInfo Team
	s.Mysql.Where(&Team{Model: gorm.Model{ID: teamID.(uint)}}).Find(&teamInfo)
	return s.makeSuccessJSON(teamInfo)
}

// 管理
func (s *Service) GetAllTeams() (int, interface{}) {
	var teams []Team
	s.Mysql.Model(&Team{}).Find(&teams)
	return s.makeSuccessJSON(teams)
}

func (s *Service) NewTeams(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		Name string `binding:"required"`
		Logo string `binding:"required"`
	}
	var inputForm []InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	// 检查传入数据是否重复
	tmpTeamName := make(map[string]int)
	for _, item := range inputForm {
		tmpTeamName[item.Name] = 0
	}
	if len(tmpTeamName) != len(inputForm){
		return s.makeErrJSON(400, 40001, "传入数据中存在重复数据")
	}

	// 与库中数据比对是否重复
	for _, item := range inputForm {
		var count int
		s.Mysql.Model(Team{}).Where(&Team{Name: item.Name}).Count(&count)
		if count != 0 {
			return s.makeErrJSON(400, 40001, "存在重复添加数据")
		}
	}

	type resultItem struct {
		Name     string
		Password string
	}
	var resultData []resultItem

	tx := s.Mysql.Begin()
	for _, item := range inputForm {
		password := randstr.String(16)
		newTeam := &Team{
			Name:      item.Name,
			Password:  s.addSalt(password),
			Logo:      item.Logo,
			SecretKey: randstr.Hex(16),
		}
		if tx.Create(newTeam).RowsAffected != 1 {
			tx.Rollback()
			return s.makeErrJSON(500, 50000, "添加 Team 失败！")
		}
		resultData = append(resultData, resultItem{
			Name:     item.Name,
			Password: password,
		})
	}
	tx.Commit()
	return s.makeSuccessJSON(resultData)
}

func (s *Service) EditTeam(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID   uint   `binding:"required"`
		Name string `binding:"required"`
		Logo string `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	// 检查 Team 是否存在
	var count int
	s.Mysql.Model(Team{}).Where(&Team{Model: gorm.Model{ID: inputForm.ID}}).Count(&count)
	if count == 0 {
		return s.makeErrJSON(404, 40400, "Team 不存在")
	}

	// 检查 Team Name 是否重复
	var repeatCheckTeam Team
	s.Mysql.Model(Team{}).Where(&Team{Name: inputForm.Name}).Find(&repeatCheckTeam)
	if repeatCheckTeam.Name != "" && repeatCheckTeam.ID != inputForm.ID {
		return s.makeErrJSON(400, 40001, "Team 重复")
	}

	newTeam := &Team{
		Name: inputForm.Name,
		Logo: inputForm.Logo,
	}
	tx := s.Mysql.Begin()
	if tx.Model(&Team{}).Where(&Team{Model: gorm.Model{ID: inputForm.ID}}).Updates(&newTeam).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50001, "修改 Team 失败！")
	}
	tx.Commit()

	return s.makeSuccessJSON("修改 Team 成功！")
}

func (s *Service) ResetTeamPassword(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID uint `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	// 检查 Team 是否存在
	var count int
	s.Mysql.Model(Team{}).Where(&Team{Model: gorm.Model{ID: inputForm.ID}}).Count(&count)
	if count == 0 {
		return s.makeErrJSON(404, 40400, "Team 不存在")
	}

	newPassword := randstr.String(16)
	tx := s.Mysql.Begin()
	if tx.Model(&Team{}).Where(&Team{Model: gorm.Model{ID: inputForm.ID}}).Updates(&Team{Password: s.addSalt(newPassword)}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50001, "重置密码失败！")
	}
	tx.Commit()

	return s.makeSuccessJSON(newPassword)
}
