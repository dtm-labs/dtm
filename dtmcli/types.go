package dtmcli

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
)

// M a short name
type M = map[string]interface{}

// MS a short name
type MS = map[string]string

// MustGenGid generate a new gid
func MustGenGid(server string) string {
	res := MS{}
	resp, err := RestyClient.R().SetResult(&res).Get(server + "/newGid")
	if err != nil || res["gid"] == "" {
		panic(fmt.Errorf("newGid error: %v, resp: %s", err, resp))
	}
	return res["gid"]
}

var sqlDbs = map[string]*sql.DB{}

// SdbGet get pooled sql.DB
func SdbGet(conf map[string]string) *sql.DB {
	dsn := GetDsn(conf)
	if sqlDbs[dsn] == nil {
		sqlDbs[dsn] = SdbAlone(conf)
	}
	return sqlDbs[dsn]
}

// SdbAlone get a standalone db connection
func SdbAlone(conf map[string]string) *sql.DB {
	dsn := GetDsn(conf)
	Logf("opening alone %s: %s", conf["driver"], strings.Replace(dsn, conf["password"], "****", 1))
	mdb, err := sql.Open(conf["driver"], dsn)
	E2P(err)
	return mdb
}

// SdbExec use raw db to exec
func SdbExec(db *sql.DB, sql string, values ...interface{}) (affected int64, rerr error) {
	r, rerr := db.Exec(sql, values...)
	if rerr == nil {
		affected, rerr = r.RowsAffected()
		Logf("affected: %d for %s %v", affected, sql, values)
	} else {
		LogRedf("exec error: %v for %s %v", rerr, sql, values)
	}
	return
}

// StxExec use raw tx to exec
func StxExec(tx *sql.Tx, sql string, values ...interface{}) (affected int64, rerr error) {
	r, rerr := tx.Exec(sql, values...)
	if rerr == nil {
		affected, rerr = r.RowsAffected()
		Logf("affected: %d for %s %v", affected, sql, values)
	} else {
		LogRedf("exec error: %v for %s %v", rerr, sql, values)
	}
	return
}

// StxQueryRow use raw tx to query row
func StxQueryRow(tx *sql.Tx, query string, args ...interface{}) *sql.Row {
	Logf("querying: "+query, args...)
	return tx.QueryRow(query, args...)
}

// GetDsn get dsn from map config
func GetDsn(conf map[string]string) string {
	conf["host"] = MayReplaceLocalhost(conf["host"])
	driver := conf["driver"]
	dsn := MS{
		"mysql": fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			conf["user"], conf["password"], conf["host"], conf["port"], conf["database"]),
		"postgres": fmt.Sprintf("host=%s user=%s password=%s dbname='%s' port=%s sslmode=disable TimeZone=Asia/Shanghai",
			conf["host"], conf["user"], conf["password"], conf["database"], conf["port"]),
	}[driver]
	PanicIf(dsn == "", fmt.Errorf("unknow driver: %s", driver))
	return dsn
}

// CheckResponse 检查Response，返回错误
func CheckResponse(resp *resty.Response, err error) error {
	if err == nil && resp != nil {
		if resp.IsError() {
			return errors.New(resp.String())
		} else if strings.Contains(resp.String(), "FAILURE") {
			return ErrFailure
		}
	}
	return err
}

// CheckResult 检查Result，返回错误
func CheckResult(res interface{}, err error) error {
	resp, ok := res.(*resty.Response)
	if ok {
		return CheckResponse(resp, err)
	}
	if res != nil && strings.Contains(MustMarshalString(res), "FAILURE") {
		return ErrFailure
	}
	return err
}

// IDGenerator used to generate a branch id
type IDGenerator struct {
	parentID string
	branchID int
}

// NewBranchID generate a branch id
func (g *IDGenerator) NewBranchID() string {
	if g.branchID >= 99 {
		panic(fmt.Errorf("branch id is larger than 99"))
	}
	if len(g.parentID) >= 20 {
		panic(fmt.Errorf("total branch id is longer than 20"))
	}
	g.branchID = g.branchID + 1
	return g.parentID + fmt.Sprintf("%02d", g.branchID)
}

// TransResult dtm 返回的结果
type TransResult struct {
	DtmResult string `json:"dtm_result"`
	Message   string
}

// TransBase 事务的基础类
type TransBase struct {
	IDGenerator
	Dtm string
	// WaitResult 是否等待全局事务的最终结果
	WaitResult bool
}

// TransBaseFromQuery construct transaction info from request
func TransBaseFromQuery(qs url.Values) *TransBase {
	return &TransBase{
		IDGenerator: IDGenerator{parentID: qs.Get("branch_id")},
		Dtm:         qs.Get("dtm"),
	}
}

// CallDtm 调用dtm服务器，返回事务的状态
func (tb *TransBase) CallDtm(body interface{}, operation string) error {
	params := MS{}
	if tb.WaitResult {
		params["wait_result"] = "1"
	}
	resp, err := RestyClient.R().SetQueryParams(params).
		SetResult(&TransResult{}).SetBody(body).Post(fmt.Sprintf("%s/%s", tb.Dtm, operation))
	if err != nil {
		return err
	}
	tr := resp.Result().(*TransResult)
	if tr.DtmResult == "FAILURE" {
		return errors.New("FAILURE: " + tr.Message)
	}
	return nil
}

// ErrFailure 表示返回失败，要求回滚
var ErrFailure = errors.New("transaction FAILURE")

// ResultSuccess 表示返回成功，可以进行下一步
var ResultSuccess = M{"dtm_result": "SUCCESS"}

// ResultFailure 表示返回失败，要求回滚
var ResultFailure = M{"dtm_result": "FAILURE"}
