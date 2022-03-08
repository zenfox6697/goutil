// test for clickhouse use
package beta

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	// _ "github.com/go-sql-driver/mysql"
	_ "github.com/mailru/go-clickhouse"
)

type SQL struct {
	// Address      string  //数据库地址:端口
	// UserName     string  //用户名
	// PassWord     string  //密码
	// DBName       string  //数据库名
	DB *sql.DB //数据库连接池对象
	// MaxOpenConns int     //用于设置最大打开的连接数，默认值为0表示不限制。
	// MaxIdleConns int     //用于设置闲置的连接数。
}

// func (m *Mysql) Init() {
// 	if m.MaxOpenConns <= 0 {
// 		m.MaxOpenConns = 30
// 	}
// 	if m.MaxIdleConns <= 0 {
// 		m.MaxIdleConns = 6
// 	}
// 	var err error
// 	m.DB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4", m.UserName, m.PassWord, m.Address, m.DBName))
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	m.DB.SetMaxOpenConns(m.MaxOpenConns)
// 	m.DB.SetMaxIdleConns(m.MaxIdleConns)
// 	m.DB.SetConnMaxLifetime(14400 * time.Second)
// 	err = m.DB.Ping()
// 	if err != nil {
// 		log.Println(err)
// 	}
// }

//新增
func (m *SQL) Insert(tableName string, data map[string]interface{}) int64 {
	return m.InsertByTx(nil, tableName, data)
}

//带有事务的新增
func (m *SQL) InsertByTx(tx *sql.Tx, tableName string, data map[string]interface{}) int64 {
	fields := []string{}
	values := []interface{}{}
	placeholders := []string{}
	if tableName == "dataexport_order" {
		if _, ok := data["user_nickname"]; ok {
			data["user_nickname"] = ""
		}
	}
	for k, v := range data {
		fields = append(fields, k)
		values = append(values, v)
		placeholders = append(placeholders, "?")
	}
	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(fields, ","), strings.Join(placeholders, ","))
	log.Println("mysql", q, values)
	return m.InsertBySqlByTx(tx, q, values...)
}

//sql语句新增
func (m *SQL) InsertBySql(q string, args ...interface{}) int64 {
	return m.InsertBySqlByTx(nil, q, args...)
}

//带有事务的sql语句新增
func (m *SQL) InsertBySqlByTx(tx *sql.Tx, q string, args ...interface{}) int64 {
	result, _ := m.ExecBySqlByTx(tx, q, args...)
	if result == nil {
		return -1
	}
	// id, err := result.LastInsertId()
	// if err != nil {
	// 	log.Println(err)
	// 	return -1
	// }
	// return id
	return 0
}

//批量新增
func (m *SQL) InsertIgnoreBatch(tableName string, fields []string, values []interface{}) (int64, int64) {
	return m.InsertIgnoreBatchByTx(nil, tableName, fields, values)
}

//带事务的批量新增
func (m *SQL) InsertIgnoreBatchByTx(tx *sql.Tx, tableName string, fields []string, values []interface{}) (int64, int64) {
	return m.insertOrReplaceBatchByTx(tx, "INSERT", "IGNORE", tableName, fields, values)
}

//批量新增
func (m *SQL) InsertBatch(tableName string, fields []string, values []interface{}) (int64, int64) {
	return m.InsertBatchByTx(nil, tableName, fields, values)
}

//带事务的批量新增
func (m *SQL) InsertBatchByTx(tx *sql.Tx, tableName string, fields []string, values []interface{}) (int64, int64) {
	return m.insertOrReplaceBatchByTx(tx, "INSERT", "", tableName, fields, values)
}

//批量更新
func (m *SQL) ReplaceBatch(tableName string, fields []string, values []interface{}) (int64, int64) {
	return m.ReplaceBatchByTx(nil, tableName, fields, values)
}

//带事务的批量更新
func (m *SQL) ReplaceBatchByTx(tx *sql.Tx, tableName string, fields []string, values []interface{}) (int64, int64) {
	return m.insertOrReplaceBatchByTx(tx, "REPLACE", "", tableName, fields, values)
}

func (m *SQL) insertOrReplaceBatchByTx(tx *sql.Tx, tp string, afterInsert, tableName string, fields []string, values []interface{}) (int64, int64) {
	placeholders := []string{}
	for range fields {
		placeholders = append(placeholders, "?")
	}
	placeholder := strings.Join(placeholders, ",")
	array := []string{}
	for i := 0; i < len(values)/len(fields); i++ {
		array = append(array, fmt.Sprintf("(%s)", placeholder))
	}
	q := fmt.Sprintf("%s %s INTO %s (%s) VALUES %s", tp, afterInsert, tableName, strings.Join(fields, ","), strings.Join(array, ","))
	result, _ := m.ExecBySqlByTx(tx, q, values...)
	if result == nil {
		return -1, -1
	}
	// v1, e1 := result.RowsAffected()
	_, e1 := result.RowsAffected()
	if e1 != nil {
		log.Println(e1)
		return -1, -1
	}
	// v2, e2 := result.LastInsertId()
	// if e2 != nil {
	// 	log.Println(e2)
	// 	return -1, -1
	// }
	// return v1, v2
	return 0, 0
}

//sql语句执行
func (m *SQL) ExecBySql(q string, args ...interface{}) (sql.Result, error) {
	return m.ExecBySqlByTx(nil, q, args...)
}

//sql语句执行,带有事务
func (m *SQL) ExecBySqlByTx(tx *sql.Tx, q string, args ...interface{}) (sql.Result, error) {
	var stmtIns *sql.Stmt
	var err error
	if tx == nil {
		stmtIns, err = m.DB.Prepare(q)
	} else {
		stmtIns, err = tx.Prepare(q)
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmtIns.Close()
	result, err := stmtIns.Exec(args...)
	if err != nil {
		log.Println(args, err)
		return nil, err
	}
	return result, nil
}

/*不等于 map[string]string{"ne":"1"}
 *不等于多个 map[string]string{"notin":[]interface{}{1,2}}
 *字段为空 map[string]string{"name":"$isNull"}
 *字段不为空 map[string]string{"name":"$isNotNull"}
 */
func (m *SQL) Find(tableName string, query map[string]interface{}, fields, order string, start, pageSize int) *[]map[string]interface{} {
	fs := []string{}
	vs := []interface{}{}
	for k, v := range query {
		rt := reflect.TypeOf(v)
		rv := reflect.ValueOf(v)
		if rt.Kind() == reflect.Map {
			for _, rv_k := range rv.MapKeys() {
				if rv_k.String() == "ne" {
					fs = append(fs, fmt.Sprintf("%s!=?", k))
					vs = append(vs, rv.MapIndex(rv_k).Interface())
				}
				if rv_k.String() == "notin" {
					if len(rv.MapIndex(rv_k).Interface().([]interface{})) > 0 {
						for _, v := range rv.MapIndex(rv_k).Interface().([]interface{}) {
							fs = append(fs, fmt.Sprintf("%s!=?", k))
							vs = append(vs, v)
						}
					}
				}
				if rv_k.String() == "in" {
					if len(rv.MapIndex(rv_k).Interface().([]interface{})) > 0 {
						_fs := fmt.Sprintf("%s in (?", k)
						for k, v := range rv.MapIndex(rv_k).Interface().([]interface{}) {
							if k > 0 {
								_fs += ",?"
							}
							vs = append(vs, v)
						}
						_fs += ")"
						fs = append(fs, _fs)
					}
				}
			}
		} else {
			if v == "$isNull" {
				fs = append(fs, fmt.Sprintf("%s is null", k))
			} else if v == "$isNotNull" {
				fs = append(fs, fmt.Sprintf("%s is not null", k))
			} else {
				fs = append(fs, fmt.Sprintf("%s=?", k))
				vs = append(vs, v)
			}
		}
	}
	var buffer bytes.Buffer
	buffer.WriteString("select ")
	if fields == "" {
		buffer.WriteString("*")
	} else {
		buffer.WriteString(fields)
	}
	buffer.WriteString(" from ")
	buffer.WriteString(tableName)
	if len(fs) > 0 {
		buffer.WriteString(" where ")
		buffer.WriteString(strings.Join(fs, " and "))
	}
	if order != "" {
		buffer.WriteString(" order by ")
		buffer.WriteString(order)
	}
	if start > -1 && pageSize > 0 {
		buffer.WriteString(" limit ")
		buffer.WriteString(fmt.Sprint(start))
		buffer.WriteString(",")
		buffer.WriteString(fmt.Sprint(pageSize))
	}
	q := buffer.String()
	log.Println(q, vs)
	return m.SelectBySql(q, vs...)
}

//sql语句查询
func (m *SQL) SelectBySql(q string, args ...interface{}) *[]map[string]interface{} {
	return m.SelectBySqlByTx(nil, q, args...)
}
func (m *SQL) SelectBySqlByTx(tx *sql.Tx, q string, args ...interface{}) *[]map[string]interface{} {
	return m.Select(0, nil, tx, q, args...)
}
func (m *SQL) Select(bath int, f func(l *[]map[string]interface{}), tx *sql.Tx, q string, args ...interface{}) *[]map[string]interface{} {
	var stmtOut *sql.Stmt
	var err error
	if tx == nil {
		stmtOut, err = m.DB.Prepare(q)
	} else {
		stmtOut, err = tx.Prepare(q)
	}
	if err != nil {
		log.Println(err)
		return nil
	}
	defer stmtOut.Close()
	rows, err := stmtOut.Query(args...)
	if err != nil {
		log.Println(err)
		return nil
	}
	if rows != nil {
		defer rows.Close()
	}
	columns, err := rows.Columns()
	if err != nil {
		log.Println(err)
		return nil
	}
	list := []map[string]interface{}{}
	for rows.Next() {
		scanArgs := make([]interface{}, len(columns))
		values := make([]interface{}, len(columns))
		ret := make(map[string]interface{})
		for k, _ := range values {
			scanArgs[k] = &values[k]
		}
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Println(err)
			break
		}
		for i, col := range values {
			if v, ok := col.([]uint8); ok {
				ret[columns[i]] = string(v)
			} else {
				ret[columns[i]] = col
			}
		}
		list = append(list, ret)
		if bath > 0 && len(list) == bath {
			f(&list)
			list = []map[string]interface{}{}
		}
	}
	if bath > 0 && len(list) > 0 {
		f(&list)
		list = []map[string]interface{}{}
	}
	return &list
}
func (m *SQL) SelectByBath(bath int, f func(l *[]map[string]interface{}), q string, args ...interface{}) {
	m.SelectByBathByTx(bath, f, nil, q, args...)
}
func (m *SQL) SelectByBathByTx(bath int, f func(l *[]map[string]interface{}), tx *sql.Tx, q string, args ...interface{}) {
	m.Select(bath, f, tx, q, args...)
}
func (m *SQL) FindOne(tableName string, query map[string]interface{}, fields, order string) *map[string]interface{} {
	list := m.Find(tableName, query, fields, order, 0, 1)
	if list != nil && len(*list) == 1 {
		temp := (*list)[0]
		return &temp
	}
	return nil
}

//修改
func (m *SQL) Update(tableName string, query, update map[string]interface{}) bool {
	return m.UpdateByTx(nil, tableName, query, update)
}

//带事务的修改
func (m *SQL) UpdateByTx(tx *sql.Tx, tableName string, query, update map[string]interface{}) bool {
	q_fs := []string{}
	u_fs := []string{}
	values := []interface{}{}
	for k, v := range update {
		q_fs = append(q_fs, fmt.Sprintf("%s=?", k))
		values = append(values, v)
	}
	for k, v := range query {
		u_fs = append(u_fs, fmt.Sprintf("%s=?", k))
		values = append(values, v)
	}
	// optmize for clickhouse
	q := fmt.Sprintf("ALTER TABLE %s UPDATE %s WHERE %s;", tableName, strings.Join(q_fs, ","), strings.Join(u_fs, " and "))
	log.Println(q, values)
	return m.UpdateOrDeleteBySqlByTx(tx, q, values...) >= 0
}

//删除
func (m *SQL) Delete(tableName string, query map[string]interface{}) bool {
	return m.DeleteByTx(nil, tableName, query)
}
func (m *SQL) DeleteByTx(tx *sql.Tx, tableName string, query map[string]interface{}) bool {
	fields := []string{}
	values := []interface{}{}
	for k, v := range query {
		fields = append(fields, fmt.Sprintf("%s=?", k))
		values = append(values, v)
	}
	// q := fmt.Sprintf("delete from %s where %s", tableName, strings.Join(fields, " and "))
	//optmize for CH
	q := fmt.Sprintf("ALTER TABLE %s DELETE WHERE %s;", tableName, strings.Join(fields, " AND "))
	log.Println(q, values)
	return m.UpdateOrDeleteBySqlByTx(tx, q, values...) > 0
}

//修改或删除
func (m *SQL) UpdateOrDeleteBySql(q string, args ...interface{}) int64 {
	return m.UpdateOrDeleteBySqlByTx(nil, q, args...)
}

//带事务的修改或删除
func (m *SQL) UpdateOrDeleteBySqlByTx(tx *sql.Tx, q string, args ...interface{}) int64 {
	result, err := m.ExecBySqlByTx(tx, q, args...)
	if err != nil {
		log.Println(err)
		return -1
	}
	count, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		return -1
	}
	return count
}

//总数
func (m *SQL) Count(tableName string, query map[string]interface{}) int64 {
	fields := []string{}
	values := []interface{}{}
	for k, v := range query {
		rt := reflect.TypeOf(v)
		rv := reflect.ValueOf(v)
		if rt.Kind() == reflect.Map {
			for _, rv_k := range rv.MapKeys() {
				if rv_k.String() == "ne" {
					fields = append(fields, fmt.Sprintf("%s!=?", k))
					values = append(values, rv.MapIndex(rv_k).Interface())
				}
				if rv_k.String() == "notin" {
					if len(rv.MapIndex(rv_k).Interface().([]interface{})) > 0 {
						for _, v := range rv.MapIndex(rv_k).Interface().([]interface{}) {
							fields = append(fields, fmt.Sprintf("%s!=?", k))
							values = append(values, v)
						}
					}
				}
				if rv_k.String() == "in" {
					if len(rv.MapIndex(rv_k).Interface().([]interface{})) > 0 {
						_fs := fmt.Sprintf("%s in (?", k)
						for k, v := range rv.MapIndex(rv_k).Interface().([]interface{}) {
							if k > 0 {
								_fs += ",?"
							}
							values = append(values, v)
						}
						_fs += ")"
						fields = append(fields, _fs)
					}
				}
			}
		} else if v == "$isNull" {
			fields = append(fields, fmt.Sprintf("%s is null", k))
		} else if v == "$isNotNull" {
			fields = append(fields, fmt.Sprintf("%s is not null", k))
		} else {
			fields = append(fields, fmt.Sprintf("%s=?", k))
			values = append(values, v)
		}
	}
	q := fmt.Sprintf("select count(1) as count from %s", tableName)
	if len(query) > 0 {
		q += fmt.Sprintf(" where %s", strings.Join(fields, " and "))
	}
	log.Println(q, values)
	return m.CountBySql(q, values...)
}
func (m *SQL) CountBySql(q string, args ...interface{}) int64 {
	stmtIns, err := m.DB.Prepare(q)
	if err != nil {
		log.Println(err)
		return -1
	}
	defer stmtIns.Close()

	rows, err := stmtIns.Query(args...)
	if err != nil {
		log.Println(err)
		return -1
	}
	if rows != nil {
		defer rows.Close()
	}
	var count int64 = -1
	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			log.Println(err)
		}
	}
	return count
}

//执行事务
func (m *SQL) ExecTx(msg string, f func(tx *sql.Tx) bool) bool {
	tx, err := m.DB.Begin()
	if err != nil {
		log.Println(msg, "获取事务错误", err)
	} else {
		if f(tx) {
			if err := tx.Commit(); err != nil {
				log.Println(msg, "提交事务错误", err)
			} else {
				return true
			}
		} else {
			if err := tx.Rollback(); err != nil {
				log.Println(msg, "事务回滚错误", err)
			}
		}
	}
	return false
}

/*************方法命名不规范，上面有替代方法*************/
func (m *SQL) Query(query string, args ...interface{}) *[]map[string]interface{} {
	return m.SelectBySql(query, args...)
}

func (m *SQL) QueryCount(query string, args ...interface{}) (count int) {
	count = -1
	if !strings.Contains(strings.ToLower(query), "count(*)") {
		fmt.Println("QueryCount need query like < select count(*) from ..... >")
		return
	}
	count = int(m.CountBySql(query, args...))
	return
}
