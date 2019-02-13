package ruleset

import(
    //"io/ioutil"
    "fmt"
    "github.com/astaxie/beego/logs"
    "bufio" //read line by line the doc
    "regexp"
    "os"
    //"strconv"
    "owlhmaster/utils"
    "owlhmaster/database"
    "errors"
    "database/sql"
)

func ReadSID(sid string)( sidLine map[string]string ,err error){

    data, err := os.Open("/etc/owlh/ruleset/owlh.rules")
    if err != nil {
        fmt.Println("File reading error", err)
        return 
    }
    
    var validID = regexp.MustCompile(`sid:`+sid+`;`)
    scanner := bufio.NewScanner(data)
    for scanner.Scan(){
        if validID.MatchString(scanner.Text()){
            sidLine := make(map[string]string)
            sidLine["raw"] = scanner.Text()
            return sidLine,err
        }    
    }
    return nil,err
}

func Read(path string)(rules map[string]map[string]string, err error) {
//leer fichero V
//dar formato a las reglas (json)
//enviar datos a ruleset.html para mostrarlos por pantalla

    //var rules map[string]string
    logs.Info ("Buscando el fichero desde ruleset/ruleset.go")
    //data, err := ioutil.ReadFile("/etc/owlh/ruleset/owlh.rules")
    //data, err := os.Open("/etc/owlh/ruleset/owlh.rules")
    data, err := os.Open(path)
    
    if err != nil {
        fmt.Println("File reading error", err)
        return 
    }

    var validID = regexp.MustCompile(`sid:(\d+);`)
    var ipfield = regexp.MustCompile(`^([^\(]+)\(`)
    var msgfield = regexp.MustCompile(`msg:([^;]+);`)
    var enablefield = regexp.MustCompile(`^#`)

    scanner := bufio.NewScanner(data)
    rules = make(map[string]map[string]string)
    for scanner.Scan(){
        if validID.MatchString(scanner.Text()){
            sid := validID.FindStringSubmatch(scanner.Text())
            msg := msgfield.FindStringSubmatch(scanner.Text())
            ip := ipfield.FindStringSubmatch(scanner.Text())
            rule := make(map[string]string)

            if enablefield.MatchString(scanner.Text()){
                rule["enabled"]="Disabled"
            }else{
                rule["enabled"]="Enabled"
            }

            rule["sid"]=sid[1]
            rule["msg"]=msg[1]
            rule["ip"]=ip[1]
            rule["raw"]=scanner.Text()
            rules[sid[1]]=rule

        }
    }
    return rules,err
}

func AddRuleset(n map[string]string) (err error) {
    logs.Info("ADD RULESET -> IN")
    //crear UUID
    rulesetID := utils.Generate()
    logs.Info(n)

    //Verificar que nos llegan los params
    if _, ok := n["name"]; !ok {
        return errors.New("Name is empty")
    }
    if _, ok := n["path"]; !ok {
        return errors.New("Path is empty")
    }
    //Verificar que no existe el ruleset
    if err := rulesetExists(rulesetID); err != nil {
        return err
    }
    //Meter en DB
    for key, value := range n {
        err = rulesetInsert(rulesetID, key, value)
    }
    if err != nil {
        return err
    }
    return nil
}

func rulesetExists(rulesetID string) (err error) {
    if ndb.Rdb == nil {
        logs.Error("rulesetExists -- Can't access to database")
        return errors.New("rulesetExists -- Can't access to database")
    }
    sql := "SELECT * FROM ruleset where ruleset_uniqueid = '"+rulesetID+"';"
    rows, err := ndb.Rdb.Query(sql)
    if err != nil {
        logs.Error(err.Error())
        return err
    }
    defer rows.Close()
    if rows.Next() {
        return errors.New("rulesetExists -- RulesetId exists")
    } else {
        return nil
    }
}

func rulesetInsert(nkey string, key string, value string) (err error) {
    if ndb.Rdb == nil {
        logs.Error("rulesetInsert -- Can't access to database")
        return errors.New("rulesetInsert -- Can't access to database")
    }
    logs.Info("nkey: %s, key: %s, value: %s", nkey, key, value)
    stmt, err := ndb.Rdb.Prepare("insert into ruleset (ruleset_uniqueid, ruleset_param, ruleset_value) values(?,?,?)")
    if err != nil {
        logs.Error("rulesetInsert -- Prepare -> %s", err.Error())
        return err
    }
    _, err = stmt.Exec(&nkey, &key, &value)
    if err != nil {
        logs.Error("rulesetInsert -- Execute -> %s", err.Error())
        return err
    }
    return nil
}

func GetAllRulesets() (rulesets *map[string]map[string]string, err error) {
    var allrulesets = map[string]map[string]string{}
    var uniqid string
    var param string
    var value string
    if ndb.Rdb == nil {
        logs.Error("ruleset/GetAllRulesets -- Can't access to database")
        return nil, errors.New("ruleset/GetAllRulesets -- Can't access to database")
    }
    sql := "select ruleset_uniqueid, ruleset_param, ruleset_value from ruleset;"
    rows, err := ndb.Rdb.Query(sql)
    if err != nil {
        logs.Error("ruleset/GetAllRulesets -- Query error: %s", err.Error())
        return nil, err
    }
    for rows.Next() {
        if err = rows.Scan(&uniqid, &param, &value); err != nil {
            logs.Error("ruleset/GetAllRulesets -- Query return error: %s", err.Error())
            return nil, err
        }
        logs.Info ("uniqid: %s, param: %s, value: %s", uniqid,param,value)
        if allrulesets[uniqid] == nil { allrulesets[uniqid] = map[string]string{}}
        allrulesets[uniqid][param]=value
    } 
    return &allrulesets, nil
}

func GetRulesetPath(nid string) (n string, err error) {
    logs.Info("DB RULESET -> Get path"+nid)
    var path string
    if ndb.Rdb != nil {
        row := ndb.Rdb.QueryRow("SELECT ruleset_value FROM ruleset WHERE ruleset_uniqueid=$1 and ruleset_param=\"path\";",nid)
        err = row.Scan(&path)
        if err == sql.ErrNoRows {
            logs.Warn("DB RULESET -> No encuentro na, ese id %s parece no existir",nid)
            return "", errors.New("DB RULESET -> No encuentro na, ese id "+nid+" parece no existir")
        }
        if err != nil {
            logs.Warn("DB RULESET -> no hemos leido bien los campos de scan")
            return "", errors.New("DB RULESET -> no hemos leido bien los campos de scan")
        }
        return path, nil
    } else {
        logs.Info("DB RULESET -> no hay base de datos")
        return "", errors.New("DB RULESET -> no hay base de datos")
    }
}

func GetRulesetRules(nid string)(r map[string]map[string]string, err error){
    logs.Info("DB RULESET -> GetRulesetRules"+nid)
    rules := make(map[string]map[string]string)
    path,err := GetRulesetPath(nid) //obtener path ruleset
    rules,err = Read(path) //obtener rules del ruleset a traves del path del parametro
    return rules, err
}