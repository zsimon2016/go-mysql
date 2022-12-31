### Golang mysql操作助手  
#### 特点
- 语意化操作
- 支持链式操作  
- 支持连接池  
- 查询结果自动转换为map
- 支持自定义查询

#### 安装
```
go get gitee.com/simon_git_code/go-mysql  
```
#### 初始化
```
    //数据库配置
	dbconf := map[string]string{
		"password": "123456",
		"username": "root",
		"network":  "tcp",
		"server":   "127.0.0.1",
		"port":     "3306",
		"database": "car",
		"prefix":   "c_",
	}
    // 获取查询实例
	Db := new(mysql.Db) 
	Db.DbConn = new(mysql.DbConnection)
	Db.DbConn.Connt(&dbconf)
	// 连接池配置(可选配置)
	// 设置与数据库建立连接的最大数目
	Db.DbConn.DB.SetMaxOpenConns(500)
	// 设置连接空闲的最大时间
	Db.DbConn.DB.SetConnMaxIdleTime(16)
	// 设置连接池中的最大闲置连接数
	Db.DbConn.DB.SetMaxIdleConns(10)
	// 设置连接可重用的最大时间
	Db.DbConn.DB.SetConnMaxLifetime(30 * time.Second)
```
#### 查询方法说明  
##### func (*DbQuery) Field  
```
func (*DbQuery) Field(field string) *DbQuery
```
语句中指定被查询的字段，支持mysql查询函数
##### func (*DbQuery) Where  
```
func (*DbQuery) Where(wher string,v interface{})  *DbQuery
```
用于Select、Delete、Update等查询条件,参数wher通常为预查询语句,例如:Where("id = ?",1),其含义为字段id的值为1的条件,?为占位符，必不可少。
###### Example
```
ret,err:=Db.Db("example").Field("*").Where("id = ?",1).Find()  
解析为查询原语：  
SELECT * FROM `example` WHERE `id`= 1 
```
#### func (*DbQuery) Or
```
func (*DbQuery) Or(or string, v interface{})
```
OR条件查询
#### func (*DbQuery) In
```
func (*DbQuery) In(in string, v []interface{}) *DbQuery
```
IN查询,参数in 字段名,v 参数数组
##### Example
```
ret,err:=Db.Db("example").In("id",[]interface{}{1,2,3}).Field("*").Select()
解析为查询语原语：
SELECT * FROM `example` WHERE `id` IN (1,2,3)
```
#### func (q *DbQuery) OrIn   
```
func (*DbQuery) OrIn(in string, v []interface{}) *DbQuery 
```
OR IN查询 参数同in查询，只是在IN查询前加OR查询
#### func (q *DbQuery) Join
```
func (*DbQuery) Join(jType string, table string, on string) *DbQuery
```
联合查询
jType -- 参数包含LEFT JOIN(left)、RIGHT JOIN(right)、INNER JOIN(inner)。 
table -- 表名，在配置中如果指明了前辍，table不需要包含前辍  
on    -- 联合查询条件
##### Example
```
Db.Db("example e").Join("left","user u","u.id = e.uid").Field("u.name,e.id").Where("id = ?",1).Find()
```
其他方法
|名称|        参数  |    说明     |
|:---|:------------|:------     |
|Order|order 字符串|ORDER BY语句|
|Group|group 字符串|GROUP BY语句|
|Size |page int 默认值1, size int 页容量|分页查询|
|Having|having 字符串 含占位符? | HAVING条件语句|
|Find|无|查询一条记录|
|Select|无|查询多条记录|
|Save|map[string]interface{}|写入数据|
|Update|无|更新数据|
|Del|无|删除记录|
|Query|字符串|自定义查询|

#### 获取写入成功后的自增ID
Save方法写入成功后，返回result实例。
```
newId,err:=result.LastInsertId()
newId即为新增的ID
```





