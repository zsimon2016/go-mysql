package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"strings"

	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type Db struct {
	DbConn *DbConnection
	// DbQuery  DbQuery
	Dbconfig *map[string]interface{}
}

//创建连接器
type DbConnection struct {
	DB       *sql.DB
	DbPrefix string
	DbName   string
	DbConf   map[string]string
}

//构建查询器
type DbQuery struct {
	Wher   string
	args   []interface{}
	Joinn  string
	order  string
	limit  string
	alias  string
	table  string
	group  string
	having string
	in     []interface{}
	field  string
	DB     *sql.DB
	Rows   interface{}
	Row    interface{}
	Prefix string
	sync.RWMutex
}

//构建连接
func (DbConnection *DbConnection) Connt(cnt *map[string]string) error {
	var build strings.Builder
	build.WriteString((*cnt)["username"])
	build.WriteString(":")
	build.WriteString((*cnt)["password"])
	build.WriteString("@")
	build.WriteString((*cnt)["network"])
	build.WriteString("(")
	build.WriteString((*cnt)["server"])
	build.WriteString(":")
	build.WriteString((*cnt)["port"])
	build.WriteString(")/")
	build.WriteString((*cnt)["database"])
	db, err := sql.Open("mysql", build.String())
	if err != nil {
		log.Printf("Open mysql failed,err:%v\n", err)
		return err
	}
	DbConnection.DB = db
	DbConnection.DbPrefix = (*cnt)["prefix"]
	DbConnection.DbName = (*cnt)["database"]
	DbConnection.DbConf = (*cnt)
	return nil
}

func (db *Db) Db(table string) *DbQuery {
	var q = new(DbQuery)
	q.Builder(db.DbConn)
	if table != "" {
		q = q.Table(table)
	}
	return q
}

func (q *DbQuery) Builder(dbConnection *DbConnection) {
	connerr := dbConnection.DB.Ping()
	if connerr != nil {
		dbconf := dbConnection.DbConf
		dbConnection.Connt(&dbconf)
		log.Printf("Mysql连接不可用 <%v>", connerr)
	}
	q.DB = dbConnection.DB
	q.Prefix = dbConnection.DbPrefix
}

func (q *DbQuery) Table(table string) *DbQuery {
	var b strings.Builder
	b.WriteString(q.Prefix)
	b.WriteString(table)
	q.table = b.String()
	return q
}
func (q *DbQuery) Alias(alias string) *DbQuery {
	q.alias = alias
	return q
}
func (q *DbQuery) In(in string, v []interface{}) *DbQuery {
	l := len(v)
	if l > 0 {
		q.in = v
		var b strings.Builder
		b.WriteString(in)
		b.WriteString(" IN (")
		for k, _ := range v {
			b.WriteString("?")
			if k < (l - 1) {
				b.WriteString(",")
			}
		}
		b.WriteString(") ")
		if q.Wher != "" {
			q.Wher += " AND " + b.String()
			q.args = append(q.args, v...)
		} else {
			q.Wher = " WHERE " + b.String()
			q.args = v
		}
	}
	return q
}
func (q *DbQuery) OrIn(in string, v []interface{}) *DbQuery {
	l := len(v)
	if l > 0 {
		q.in = v
		var b strings.Builder
		b.WriteString(in)
		b.WriteString(" IN (")
		for k, _ := range v {
			b.WriteString("?")
			if k < (l - 1) {
				b.WriteString(",")
			}
		}
		b.WriteString(") ")
		if q.Wher != "" {
			q.Wher += " OR " + b.String()
			q.args = append(q.args, v...)
		} else {
			q.Wher = " WHERE " + b.String()
			q.args = v
		}
	}
	return q
}
func (q *DbQuery) Where(wher string, v interface{}) *DbQuery {
	if wher != "" {
		if q.Wher != "" {
			q.Wher += " AND " + wher
			q.args = append(q.args, v)
		} else {
			q.Wher = " WHERE " + wher
			q.args = append(q.args, v)
		}
	}
	return q
}
func (q *DbQuery) Or(or string, v interface{}) *DbQuery {
	if or != "" {
		if q.Wher != "" {
			q.Wher += " OR " + or
			q.args = append(q.args, v)
		} else {
			q.Wher = " Where " + or
			q.args = append(q.args, v)
		}
	}
	return q
}
func (q *DbQuery) Order(order string) *DbQuery {
	if order != "" {
		q.order = " ORDER BY " + order
	}
	return q
}

//涉及线程安全，改用字符串
func (q *DbQuery) Join(jType string, table string, on string) *DbQuery {
	join := ""
	if q.Joinn != "" {
		join = q.Joinn
	}
	var b strings.Builder
	b.WriteString(join)
	joinStr := ""
	switch jType {
	case "left":
		joinStr = " LEFT JOIN "
	case "right":
		joinStr = " RIGHT JOIN "
	case "inner":
		joinStr = " INNER JOIN "
	}
	b.WriteString(joinStr)
	b.WriteString(q.Prefix)
	b.WriteString(table)
	b.WriteString(" ON ")
	b.WriteString(on)
	b.WriteString(" ")
	q.Joinn = b.String()
	return q
}

func (q *DbQuery) Count() int {
	q.Field("count(*) as count")
	query := q.SelectSql()

	log.Println(query, q.args)
	rows, err := q.DB.Query(query, q.args...)
	if err != nil {
		log.Println("Query error", err)
	}
	ret := q.GetRow(rows)["count"]
	switch ret.(type) {
	case int64:
		return int(ret.(int64))
	case string:
		res, _ := strconv.Atoi(ret.(string))
		return res
	}
	return 0
}

func (q *DbQuery) Field(field string) *DbQuery {
	if field != "" {
		q.field = field
	}
	return q
}

func (q *DbQuery) SelectSql() string {
	var b strings.Builder
	b.WriteString("SELECT ")
	b.WriteString(q.field)
	b.WriteString(" FROM ")
	b.WriteString(q.table)
	b.WriteString(" ")
	b.WriteString(q.alias)
	b.WriteString(" ")
	b.WriteString(q.Joinn)
	b.WriteString(" ")
	b.WriteString(q.Wher)
	b.WriteString(" ")
	b.WriteString(q.group)
	b.WriteString(" ")
	b.WriteString(q.having)
	b.WriteString(" ")
	b.WriteString(q.order)
	b.WriteString(" ")
	b.WriteString(q.limit)
	return b.String()
}
func (q *DbQuery) Query(sql string, args ...interface{}) ([]map[string]interface{}, error) {
	log.Println(sql, args)
	ctx := context.Background()
	conn, e := q.DB.Conn(ctx)
	if e != nil {
		conn.Close()
		log.Println("Query error", e)
		return nil, e
	}
	rows, err := conn.QueryContext(ctx, sql, args...)
	if err != nil {
		conn.Close()
		log.Println("Query error", err)
		return nil, err
	}
	defer func() {
		conn.Close()
	}()
	return q.GetRows(rows), nil
}

func (q *DbQuery) Exec(sql string, args ...interface{}) (sql.Result, error) {
	log.Println(sql, args)
	ctx := context.Background()
	conn, e := q.DB.Conn(ctx)
	if e != nil {
		conn.Close()
		log.Println("Query error", e)
		return nil, e
	}
	res, err := conn.ExecContext(ctx, sql, args...)
	if err != nil {
		conn.Close()
		log.Println("Query error", err)
		return nil, err
	}
	defer func() {
		conn.Close()
	}()
	return res, nil
}

func (q *DbQuery) Select() ([]map[string]interface{}, error) {
	query := q.SelectSql()
	log.Println(query, q.args)
	ctx := context.Background()
	conn, e := q.DB.Conn(ctx)
	if e != nil {
		conn.Close()
		log.Println("Query error", e)
		return nil, e
	}
	rows, err := conn.QueryContext(ctx, query, q.args...)
	if err != nil {
		conn.Close()
		log.Println("Query error", err)
		return nil, err
	}
	defer func() {
		conn.Close()
	}()
	return q.GetRows(rows), nil
}
func (q *DbQuery) Find() (map[string]interface{}, error) {
	query := q.SelectSql()
	log.Println(query, q.args)
	//使用DB查询
	var build strings.Builder
	build.WriteString(query)
	build.WriteString(" LIMIT 1")
	rows, err := q.DB.Query(build.String(), q.args...)
	if err != nil {
		log.Println("Query error", err)
		return nil, err
	}

	return q.GetRow(rows), nil
}
func (q *DbQuery) Update(update map[string]interface{}) (sql.Result, error) {
	defer func() {
		if err := recover(); err != nil {
			stack := debug.Stack()
			log.Println(err, string(stack))
		}
	}()
	query, value := buileUpdateParams(q, update, q.args...)
	if query == "" {
		return nil, nil
	}
	fmt.Println(query, update, value, q.args)
	ctx, cancel := context.WithCancel(context.Background())
	conn, e := q.DB.Conn(ctx)
	if e != nil {
		conn.Close()
		cancel()
		log.Println("Query error", e)
		return nil, e
	}
	res, err := conn.ExecContext(ctx, query, value...)
	if err != nil {
		conn.Close()
		cancel()
		log.Println("Query error", err)
		return nil, err
	}
	defer func() {
		conn.Close()
		cancel()
	}()
	return res, nil
}

func (q *DbQuery) Save(save map[string]interface{}) (sql.Result, error) {
	query, value := buileSaveParams(q, save)
	if query == "" {
		return nil, nil
	}
	log.Println(query, save)
	ctx := context.Background()
	conn, e := q.DB.Conn(ctx)
	if e != nil {
		log.Println("Query error", e)
		return nil, e
	}
	res, err := conn.ExecContext(ctx, query, value...)
	if err != nil {
		conn.Close()
		log.Println("Query error", err)
		return nil, err
	}
	conn.Close()
	return res, nil
}

func buileUpdateParams(q *DbQuery, update map[string]interface{}, where ...interface{}) (string, []interface{}) {
	into := ""
	var value []interface{}
	var binto strings.Builder
	var b strings.Builder
	if update != nil {
		i := len(update)
		for key := range update {
			i--
			binto.WriteString(key)
			binto.WriteString(" = ?")
			if len(update) > 0 && i > 0 {
				binto.WriteString(",")
			}
			value = append(value, update[key])
		}
		into = binto.String()
		b.WriteString("UPDATE ")
		b.WriteString(q.table)
		b.WriteString(" SET ")
		b.WriteString(into)
		b.WriteString(" ")
		b.WriteString(q.Wher)
	}
	if len(where) > 0 {
		for _, v := range where {
			value = append(value, v)
		}
	}
	query := b.String()
	return query, value
}
func buileSaveParams(q *DbQuery, save map[string]interface{}) (string, []interface{}) {
	query, into := "", ""

	var value []interface{}
	var buildPlace strings.Builder
	var buildInto strings.Builder
	var b strings.Builder

	if save != nil {
		buildInto.WriteString(" (")
		buildPlace.WriteString(")VALUES( ")
		i := len(save)
		for key := range save {
			i--
			buildInto.WriteString(key)
			buildPlace.WriteString("?")
			if len(save) > 0 && i > 0 {
				buildInto.WriteString(",")
				buildPlace.WriteString(",")
			}
			value = append(value, save[key])
		}
		place := buildPlace.String()
		into = buildInto.String()
		b.WriteString("INSERT INTO ")
		b.WriteString(q.table)
		b.WriteString(into)
		b.WriteString(place)
		b.WriteString(")")
		query = b.String()
		return query, value
	} else {
		return query, nil
	}
}

//删除
func buileDeleteParams(q *DbQuery) string {
	var b strings.Builder
	b.WriteString("DELETE FROM ")
	b.WriteString(q.table)
	b.WriteString(" ")
	b.WriteString(q.Wher)
	query := b.String()
	return query
}

//创建事务
func (q *DbQuery) CreateDBTx() (context.Context, *sql.Tx, error) {
	ctx := context.Background()
	conn, e := q.DB.Conn(ctx)
	if e == nil {
		opts := new(sql.TxOptions)
		tx, err := conn.BeginTx(ctx, opts)
		if err == nil {
			return ctx, tx, err
		} else {
			return nil, nil, err
		}

	}
	log.Println("===创建事务失败:===", e)
	return nil, nil, e
}

//事务写入
func (q *DbQuery) TxSave(ctx context.Context, tx *sql.Tx, save map[string]interface{}) (sql.Result, error) {
	query, value := buileSaveParams(q, save)
	if query == "" {
		return nil, nil
	}
	log.Println(query, value)
	var result sql.Result
	res, err := tx.ExecContext(ctx, query, value...)
	if err != nil {
		log.Println("Query error", err)
		e := tx.Rollback()
		if e != nil {
			log.Println("Query error", e)
		}
	}
	result = res
	return result, err
}

//事务更新
func (q *DbQuery) TxUpdate(ctx context.Context, tx *sql.Tx, update map[string]interface{}, where ...interface{}) (sql.Result, error) {
	query, value := buileUpdateParams(q, update, where...)
	if query == "" {
		return nil, nil
	}
	log.Println(query, value)
	var result sql.Result
	res, err := tx.ExecContext(ctx, query, value...)
	if err != nil {
		log.Println("Query error", err)
		e := tx.Rollback()
		if e != nil {
			log.Println("Query error", e)
		}
	}
	result = res
	return result, err
}

func (q *DbQuery) Del() sql.Result {
	query := buileDeleteParams(q)
	if query == "" {
		return nil
	}
	log.Println(query, q.args)
	ctx := context.Background()
	conn, e := q.DB.Conn(ctx)
	var result sql.Result
	if e == nil {
		res, err := conn.ExecContext(ctx, query, q.args...)
		if err != nil {
			log.Println("Query error", err)
		}
		result = res
		conn.Close()
	} else {
		log.Println("Query error", e)
	}
	conn.Close()
	return result
}
func (q *DbQuery) Limit(limit string) *DbQuery {
	if limit != "" {
		q.limit = " LIMIT " + limit + ""
	}
	return q
}

func (q *DbQuery) Size(page int, size int) *DbQuery {
	start := (page - 1) * size
	startToStr := strconv.Itoa(start)
	sizeToStr := strconv.Itoa(size)
	var b strings.Builder
	b.WriteString(" LIMIT ")
	b.WriteString(startToStr)
	b.WriteString(",")
	b.WriteString(sizeToStr)
	limitStr := b.String()
	q.limit = limitStr
	return q
}

func (q *DbQuery) Group(group string) *DbQuery {
	if group != "" {
		q.group = " GROUP BY " + group
	}
	return q
}
func (q *DbQuery) Having(h string, v interface{}) *DbQuery {
	if h != "" {
		q.having = " HAVING " + h
		q.args = append(q.args, v)
	}
	return q
}
func (q *DbQuery) GetRows(query *sql.Rows) []map[string]interface{} {
	defer func() {
		query.Close()
	}()
	columns, _ := query.Columns()
	field := make([]interface{}, len(columns))
	//让每一行数据都填充到[][]byte里面
	for i := range field {
		var v interface{}
		field[i] = &v
	}
	var results []map[string]interface{}
	for query.Next() {
		if err := query.Scan(field...); err != nil {
			fmt.Println(err)
		}
		row := make(map[string]interface{})
		for k, val := range field {
			v := *(val.(*interface{}))
			key := columns[k]
			row[key] = assertion(v)
		}
		results = append(results, row)
	}
	return results
}
func (q *DbQuery) GetRow(query *sql.Rows) map[string]interface{} {
	defer func() {
		query.Close()
	}()
	columns, _ := query.Columns()
	field := make([]interface{}, len(columns))
	for i := range field {
		var v interface{}
		field[i] = &v
	}
	row := map[string]interface{}{}
	for query.Next() {
		if err := query.Scan(field...); err != nil {
			fmt.Println(err)
		}
		for k, val := range field {
			v := *(val.(*interface{}))
			key := columns[k]
			row[key] = assertion(v)
		}
		break
	}
	return row
}
func assertion(v interface{}) interface{} {
	var t interface{}
	switch v.(type) {
	case int64:
		t = int64(v.(int64))
	case int32:
		t = int32(v.(int32))
	case int16:
		t = int16(v.(int16))
	case int8:
		t = int8(v.(int8))
	case int:
		t = int(v.(int))
	case []uint8:
		t = string(v.([]uint8))
	case float32:
		t = float32(v.(float32))
	case float64:
		t = float64(v.(float64))
	case uint8:
		t = uint8(v.(uint8))
	case uint16:
		t = uint16(v.(uint16))
	case uint32:
		t = uint32(v.(uint32))
	case uint64:
		t = uint64(v.(uint64))
	case nil:
		t = nil
	default:
		t = v
	}
	return t
}
