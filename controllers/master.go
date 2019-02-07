package controllers

import (
	"owlhmaster/models"
	"encoding/json"

	"github.com/astaxie/beego"
    "github.com/astaxie/beego/logs"
)

type MasterController struct {
	beego.Controller
}

// @Title CreateMaster
// @Description create master
// @Success 200 {int} models.master.id
// @Failure 403 body is empty
// @router / [post]
func (m *MasterController) Post() {
	mid := models.InitMaster()
	m.Data["json"] = map[string]string{"mid": mid}
	m.ServeJSON()
}


// @Title GetMaster
// @Description get Master detail
// @Success 200 {object} models.Master
// @router / [get]
func (m *MasterController) Get() {
    logs.Info ("Master controller -> GET")
    mm, err := models.GetMaster()
    if err != nil {
        logs.Info ("Master Controller -> Get -> Error %s", err)
    }
    logs.Info ("Master id es %s", mm)
    m.Data["json"] = mm
    logs.Info ("Master Detail es %s", m)
    m.ServeJSON()
}

// @Title Update
// @Description update the master
// @Param       body            body    models.Master     true            "body for master content"
// @Success 200 {object} models.Master
// @router / [put]
func (m *MasterController) Put() {
    var mm map[string]string
    json.Unmarshal(m.Ctx.Input.RequestBody, &mm)
    err := models.UpdateMaster(mm)
    m.Data["json"] = map[string]string{"ack": "true"}
    if err != nil {
        m.Data["json"] = map[string]string{"ack": "false: " + err.Error()}
    }
    m.ServeJSON()
}

