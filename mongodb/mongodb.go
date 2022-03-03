package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMgo(addr, db string, size int) *MongodbSim {
	mgo := &MongodbSim{
		MongodbAddr: addr,
		Size:        size,
		DbName:      db,
	}
	mgo.InitPool()
	return mgo
}

func NewMgoWithUser(addr, db, uname, upwd string, size int) *MongodbSim {
	mgo := &MongodbSim{
		MongodbAddr: addr,
		Size:        size,
		DbName:      db,
		UserName:    uname,
		Password:    upwd,
	}
	mgo.InitPool()
	return mgo
}

type Bluk struct {
	ms     *MgoSess
	writes []mongo.WriteModel
}

func (b *Bluk) Insert(doc interface{}) {
	write := mongo.NewInsertOneModel()
	write.SetDocument(doc)
	b.writes = append(b.writes, write)
}
func (b *Bluk) Update(doc ...interface{}) {
	write := mongo.NewUpdateOneModel()
	write.SetFilter(doc[0])
	ue := ObjToM(doc[1])
	autoUpdateTime(b.ms.db, b.ms.coll, ue)
	write.SetUpdate(ue)
	write.SetUpsert(false)
	b.writes = append(b.writes, write)
}
func (b *Bluk) UpdateAll(doc ...interface{}) {
	write := mongo.NewUpdateManyModel()
	write.SetFilter(doc[0])
	ue := ObjToM(doc[1])
	autoUpdateTime(b.ms.db, b.ms.coll, ue)
	write.SetUpdate(ue)
	write.SetUpsert(false)
	b.writes = append(b.writes, write)
}
func (b *Bluk) Upsert(doc ...interface{}) {
	write := mongo.NewUpdateOneModel()
	write.SetFilter(doc[0])
	ue := ObjToM(doc[1])
	autoUpdateTime(b.ms.db, b.ms.coll, ue)
	write.SetUpdate(ue)
	write.SetUpsert(true)
	b.writes = append(b.writes, write)
}
func (b *Bluk) Remove(doc interface{}) {
	write := mongo.NewDeleteOneModel()
	write.SetFilter(doc)
	b.writes = append(b.writes, write)
}
func (b *Bluk) RemoveAll(doc interface{}) {
	write := mongo.NewDeleteManyModel()
	write.SetFilter(doc)
	b.writes = append(b.writes, write)
}
func (b *Bluk) Run() (*mongo.BulkWriteResult, error) {
	return b.ms.M.C.Database(b.ms.db).Collection(b.ms.coll).BulkWrite(b.ms.M.Ctx, b.writes)
}

//
type MgoIter struct {
	Cursor *mongo.Cursor
	Ctx    context.Context
}

func (mt *MgoIter) Next(result interface{}) bool {
	if mt.Cursor != nil {
		if mt.Cursor.Next(mt.Ctx) {
			rType := reflect.TypeOf(result)
			rVal := reflect.ValueOf(result)
			if rType.Kind() == reflect.Ptr {
				rType = rType.Elem()
				rVal = rVal.Elem()
			}
			var err error
			if rType.Kind() == reflect.Map {
				r := make(map[string]interface{})
				err = mt.Cursor.Decode(&r)
				if rVal.CanSet() {
					rVal.Set(reflect.ValueOf(r))
				} else {
					for it := rVal.MapRange(); it.Next(); {
						rVal.SetMapIndex(it.Key(), reflect.Value{})
					}
					for it := reflect.ValueOf(r).MapRange(); it.Next(); {
						rVal.SetMapIndex(it.Key(), it.Value())
					}
				}
			} else {
				err = mt.Cursor.Decode(&result)
			}
			if err != nil {
				log.Println("mgo cur err", err.Error())
				mt.Cursor.Close(mt.Ctx)
				return false
			}
			return true
		} else {
			mt.Cursor.Close(mt.Ctx)
			return false
		}
	} else {
		return false
	}
}

//
type MgoSess struct {
	db     string
	coll   string
	query  interface{}
	sorts  []string
	fields interface{}
	limit  int64
	skip   int64
	pipe   []map[string]interface{}
	all    interface{}
	M      *MongodbSim
}

func (ms *MgoSess) DB(name string) *MgoSess {
	ms.db = name
	return ms
}
func (ms *MgoSess) C(name string) *MgoSess {
	ms.coll = name
	return ms
}
func (ms *MgoSess) Bulk() *Bluk {
	return &Bluk{ms: ms}
}
func (ms *MgoSess) Find(q interface{}) *MgoSess {
	if q == nil {
		q = map[string]interface{}{}
	}
	ms.query = q
	return ms
}
func (ms *MgoSess) FindId(_id interface{}) *MgoSess {
	ms.query = map[string]interface{}{"_id": _id}
	return ms
}
func (ms *MgoSess) Select(fields interface{}) *MgoSess {
	ms.fields = fields
	return ms
}
func (ms *MgoSess) Limit(limit int64) *MgoSess {
	ms.limit = limit
	return ms
}
func (ms *MgoSess) Skip(skip int64) *MgoSess {
	ms.skip = skip
	return ms
}
func (ms *MgoSess) Sort(sorts ...string) *MgoSess {
	ms.sorts = sorts
	return ms
}
func (ms *MgoSess) Pipe(p []map[string]interface{}) *MgoSess {
	ms.pipe = p
	return ms
}
func (ms *MgoSess) Insert(doc interface{}) error {
	_, err := ms.M.C.Database(ms.db).Collection(ms.coll).InsertOne(ms.M.Ctx, doc)
	return err
}
func (ms *MgoSess) Remove(filter interface{}) error {
	_, err := ms.M.C.Database(ms.db).Collection(ms.coll).DeleteOne(ms.M.Ctx, filter)
	return err
}
func (ms *MgoSess) RemoveId(_id interface{}) error {
	_, err := ms.M.C.Database(ms.db).Collection(ms.coll).DeleteOne(ms.M.Ctx, map[string]interface{}{"_id": _id})
	return err
}
func (ms *MgoSess) RemoveAll(filter interface{}) (*mongo.DeleteResult, error) {
	return ms.M.C.Database(ms.db).Collection(ms.coll).DeleteMany(ms.M.Ctx, filter)
}
func (ms *MgoSess) Upsert(filter, update interface{}) (*mongo.UpdateResult, error) {
	ct := options.Update()
	ct.SetUpsert(true)
	ue := ObjToM(update)
	autoUpdateTime(ms.db, ms.coll, ue)
	return ms.M.C.Database(ms.db).Collection(ms.coll).UpdateOne(ms.M.Ctx, filter, ue, ct)
}
func (ms *MgoSess) UpsertId(filter, update interface{}) (*mongo.UpdateResult, error) {
	ct := options.Update()
	ct.SetUpsert(true)
	ue := ObjToM(update)
	autoUpdateTime(ms.db, ms.coll, ue)
	return ms.M.C.Database(ms.db).Collection(ms.coll).UpdateOne(ms.M.Ctx, map[string]interface{}{"_id": filter}, ue, ct)
}
func (ms *MgoSess) UpdateId(filter, update interface{}) error {
	ue := ObjToM(update)
	autoUpdateTime(ms.db, ms.coll, ue)
	_, err := ms.M.C.Database(ms.db).Collection(ms.coll).UpdateOne(ms.M.Ctx, map[string]interface{}{"_id": filter}, ue)
	return err
}
func (ms *MgoSess) Update(filter, update interface{}) error {
	ue := ObjToM(update)
	autoUpdateTime(ms.db, ms.coll, ue)
	_, err := ms.M.C.Database(ms.db).Collection(ms.coll).UpdateOne(ms.M.Ctx, filter, ue)
	return err
}
func (ms *MgoSess) Count() (int64, error) {
	return ms.M.C.Database(ms.db).Collection(ms.coll).CountDocuments(ms.M.Ctx, ms.query)
}
func (ms *MgoSess) One(v *map[string]interface{}) {
	of := options.FindOne()
	of.SetProjection(ms.fields)
	sr := ms.M.C.Database(ms.db).Collection(ms.coll).FindOne(ms.M.Ctx, ms.query, of)
	if sr.Err() == nil {
		sr.Decode(&v)
	}
}
func (ms *MgoSess) All(v *[]map[string]interface{}) {
	cur, err := ms.M.C.Database(ms.db).Collection(ms.coll).Aggregate(ms.M.Ctx, ms.pipe)
	if err == nil && cur.Err() == nil {
		cur.All(ms.M.Ctx, v)
	}
}
func (ms *MgoSess) Iter() *MgoIter {
	it := &MgoIter{}
	coll := ms.M.C.Database(ms.db).Collection(ms.coll)
	var cur *mongo.Cursor
	var err error
	if ms.query != nil {
		find := options.Find()
		if ms.skip > 0 {
			find.SetSkip(ms.skip)
		}
		if ms.limit > 0 {
			find.SetLimit(ms.limit)
		}
		find.SetBatchSize(100)
		if len(ms.sorts) > 0 {
			sort := bson.D{}
			for _, k := range ms.sorts {
				switch k[:1] {
				case "-":
					sort = append(sort, bson.E{k[1:], -1})
				case "+":
					sort = append(sort, bson.E{k[1:], 1})
				default:
					sort = append(sort, bson.E{k, 1})
				}
			}
			find.SetSort(sort)
		}
		if ms.fields != nil {
			find.SetProjection(ms.fields)
		}
		cur, err = coll.Find(ms.M.Ctx, ms.query, find)
		if err != nil {
			log.Println("mgo find err", err.Error())
		}
	} else if ms.pipe != nil {
		aggregate := options.Aggregate()
		aggregate.SetBatchSize(100)
		cur, err = coll.Aggregate(ms.M.Ctx, ms.pipe, aggregate)
		if err != nil {
			log.Println("mgo aggregate err", err.Error())
		}
	}
	if err == nil {
		it.Cursor = cur
		it.Ctx = ms.M.Ctx
	}
	return it
}

type MongodbSim struct {
	MongodbAddr string
	Size        int
	//	MinSize     int
	DbName   string
	C        *mongo.Client
	Ctx      context.Context
	ShortCtx context.Context
	pool     chan bool
	UserName string
	Password string
	ReplSet  string
}

func (m *MongodbSim) GetMgoConn() *MgoSess {
	//m.Open()
	ms := &MgoSess{}
	ms.M = m
	return ms
}

func (m *MongodbSim) DestoryMongoConn(ms *MgoSess) {
	//m.Close()
	ms.M = nil
	ms = nil
}

func (m *MongodbSim) Destroy() {
	//m.Close()
	m.C.Disconnect(nil)
	m.C = nil
}

func (m *MongodbSim) InitPool() {
	opts := options.Client()
	registry := bson.NewRegistryBuilder().RegisterTypeMapEntry(bson.TypeArray, reflect.TypeOf([]interface{}{})).Build()
	opts.SetRegistry(registry)
	opts.SetConnectTimeout(3 * time.Second)
	opts.SetHosts(strings.Split(m.MongodbAddr, ","))
	//opts.ApplyURI("mongodb://" + m.MongodbAddr)
	opts.SetMaxPoolSize(uint64(m.Size))
	if m.UserName != "" && m.Password != "" {
		cre := options.Credential{
			Username: m.UserName,
			Password: m.Password,
		}
		opts.SetAuth(cre)
	}
	/*ms := strings.Split(m.MongodbAddr, ",")
	if m.ReplSet == "" && len(ms) > 1 {
		m.ReplSet = "qfws"
	}*/
	if m.ReplSet != "" {
		opts.SetReplicaSet(m.ReplSet)
		opts.SetDirect(false)
	}
	m.pool = make(chan bool, m.Size)
	opts.SetMaxConnIdleTime(2 * time.Hour)
	m.Ctx, _ = context.WithTimeout(context.Background(), 99999*time.Hour)
	m.ShortCtx, _ = context.WithTimeout(context.Background(), 1*time.Minute)
	client, err := mongo.Connect(m.ShortCtx, opts)
	if err != nil {
		log.Println("mgo init error:", err.Error())
	} else {
		m.C = client
	}
}

func (m *MongodbSim) Open() {
	m.pool <- true
}
func (m *MongodbSim) Close() {
	<-m.pool
}

func (m *MongodbSim) Save(c string, doc interface{}) string {
	defer catch()
	m.Open()
	defer m.Close()
	coll := m.C.Database(m.DbName).Collection(c)
	obj := ObjToM(doc)
	id := primitive.NewObjectID()
	(*obj)["_id"] = id
	_, err := coll.InsertOne(m.Ctx, obj)
	if nil != err {
		log.Println("SaveError", err)
		return ""
	}
	return id.Hex()
}

//原_id不变
func (m *MongodbSim) SaveByOriID(c string, doc interface{}) bool {
	defer catch()
	m.Open()
	defer m.Close()
	coll := m.C.Database(m.DbName).Collection(c)
	_, err := coll.InsertOne(m.Ctx, ObjToM(doc))
	if nil != err {
		log.Println("SaveByOriIDError", err)
		return false
	}
	return true
}

//批量插入
func (m *MongodbSim) SaveBulk(c string, doc ...map[string]interface{}) bool {
	defer catch()
	m.Open()
	defer m.Close()
	coll := m.C.Database(m.DbName).Collection(c)
	var writes []mongo.WriteModel
	for _, d := range doc {
		write := mongo.NewInsertOneModel()
		write.SetDocument(d)
		writes = append(writes, write)
	}
	br, e := coll.BulkWrite(m.Ctx, writes)
	if e != nil {
		b := strings.Index(e.Error(), "duplicate") > -1
		log.Println("mgo savebulk error:", e.Error())
		if br != nil {
			log.Println("mgo savebulk size:", br.InsertedCount)
		}
		return b
	}
	return true
}

//批量插入
func (m *MongodbSim) SaveBulkInterface(c string, doc ...interface{}) bool {
	defer catch()
	m.Open()
	defer m.Close()
	coll := m.C.Database(m.DbName).Collection(c)
	var writes []mongo.WriteModel
	for _, d := range doc {
		write := mongo.NewInsertOneModel()
		write.SetDocument(d)
		writes = append(writes, write)
	}
	br, e := coll.BulkWrite(m.Ctx, writes)
	if e != nil {
		b := strings.Index(e.Error(), "duplicate") > -1
		log.Println("mgo SaveBulkInterface error:", e.Error())
		if br != nil {
			log.Println("mgo SaveBulkInterface size:", br.InsertedCount)
		}
		return b
	}
	return true
}

//按条件统计
func (m *MongodbSim) Count(c string, q interface{}) int {
	r, _ := m.CountByErr(c, q)
	return r
}

//统计
func (m *MongodbSim) CountByErr(c string, q interface{}) (int, error) {
	defer catch()
	m.Open()
	defer m.Close()
	var res int64
	var err error
	if filter := ObjToM(q); filter != nil && len(*filter) > 0 {
		res, err = m.C.Database(m.DbName).Collection(c).CountDocuments(m.Ctx, filter)
	} else {
		res, err = m.C.Database(m.DbName).Collection(c).EstimatedDocumentCount(m.Ctx)
	}
	if err != nil {
		log.Println("统计错误", err.Error())
		return 0, err
	} else {
		return int(res), nil
	}
}

//按条件删除
func (m *MongodbSim) Delete(c string, q interface{}) int64 {
	defer catch()
	m.Open()
	defer m.Close()
	res, err := m.C.Database(m.DbName).Collection(c).DeleteMany(m.Ctx, ObjToM(q))
	if err != nil && res == nil {
		log.Println("删除错误", err.Error())
	}
	return res.DeletedCount
}

//删除对象
func (m *MongodbSim) Del(c string, q interface{}) bool {
	defer catch()
	m.Open()
	defer m.Close()
	_, err := m.C.Database(m.DbName).Collection(c).DeleteMany(m.Ctx, ObjToM(q))
	if err != nil {
		log.Println("删除错误", err.Error())
		return false
	}
	return true
}

//按条件更新
func (m *MongodbSim) Update(c string, q, u interface{}, upsert bool, multi bool) bool {
	defer catch()
	m.Open()
	defer m.Close()
	ct := options.Update()
	if upsert {
		ct.SetUpsert(true)
	}
	coll := m.C.Database(m.DbName).Collection(c)
	ue := ObjToM(u)
	autoUpdateTime(m.DbName, c, ue)
	var err error
	if multi {
		_, err = coll.UpdateMany(m.Ctx, ObjToM(q), ue, ct)
	} else {
		_, err = coll.UpdateOne(m.Ctx, ObjToM(q), ue, ct)
	}
	if err != nil {
		log.Println("更新错误", err.Error())
		return false
	}
	return true
}
func (m *MongodbSim) UpdateById(c string, id interface{}, set interface{}) bool {
	defer catch()
	m.Open()
	defer m.Close()
	q := make(map[string]interface{})
	if sid, ok := id.(string); ok {
		q["_id"], _ = primitive.ObjectIDFromHex(sid)
	} else {
		q["_id"] = id
	}
	ue := ObjToM(set)
	autoUpdateTime(m.DbName, c, ue)
	_, err := m.C.Database(m.DbName).Collection(c).UpdateOne(m.Ctx, q, ue)
	if nil != err {
		log.Println("UpdateByIdError", err)
		return false
	}
	return true
}

//批量更新
func (m *MongodbSim) UpdateBulkAll(db, c string, doc ...[]map[string]interface{}) bool {
	return m.NewUpdateBulk(db, c, false, false, doc...)
}

func (m *MongodbSim) UpdateBulk(c string, doc ...[]map[string]interface{}) bool {
	return m.UpdateBulkAll(m.DbName, c, doc...)
}

//批量插入
func (m *MongodbSim) UpSertBulk(c string, doc ...[]map[string]interface{}) bool {
	return m.NewUpdateBulk(m.DbName, c, true, false, doc...)
}

//批量插入
func (m *MongodbSim) UpSertMultiBulk(c string, upsert, multi bool, doc ...[]map[string]interface{}) bool {
	return m.NewUpdateBulk(m.DbName, c, upsert, multi, doc...)
}

//批量插入
func (m *MongodbSim) NewUpdateBulk(db, c string, upsert, multi bool, doc ...[]map[string]interface{}) bool {
	defer catch()
	m.Open()
	defer m.Close()
	coll := m.C.Database(db).Collection(c)
	var writes []mongo.WriteModel
	for _, d := range doc {
		if multi {
			write := mongo.NewUpdateManyModel()
			write.SetFilter(d[0])
			ue := ObjToM(d[1])
			autoUpdateTime(m.DbName, c, ue)
			write.SetUpdate(ue)
			write.SetUpsert(upsert)
			writes = append(writes, write)
		} else {
			write := mongo.NewUpdateOneModel()
			write.SetFilter(d[0])
			ue := ObjToM(d[1])
			autoUpdateTime(m.DbName, c, ue)
			write.SetUpdate(ue)
			write.SetUpsert(upsert)
			writes = append(writes, write)
		}
	}
	br, e := coll.BulkWrite(m.Ctx, writes)
	if e != nil {
		log.Println("mgo upsert error:", e.Error())
		return br == nil || br.UpsertedCount == 0
	}
	//	else {
	//		if r.UpsertedCount != int64(len(doc)) {
	//			log.Println("mgo upsert uncomplete:uc/dc", r.UpsertedCount, len(doc))
	//		}
	//		return true
	//	}
	return true
}

//查询单条对象
func (m *MongodbSim) FindOne(c string, query interface{}) (*map[string]interface{}, bool) {
	return m.FindOneByField(c, query, nil)
}

//查询单条对象
func (m *MongodbSim) FindOneByField(c string, query interface{}, fields interface{}) (*map[string]interface{}, bool) {
	defer catch()
	res, ok := m.Find(c, query, nil, fields, true, -1, -1)
	if nil != res && len(*res) > 0 {
		return &((*res)[0]), ok
	}
	return nil, ok
}

//查询单条对象
func (m *MongodbSim) FindById(c string, query string, fields interface{}) (*map[string]interface{}, bool) {
	defer catch()
	m.Open()
	defer m.Close()
	of := options.FindOne()
	of.SetProjection(ObjToOth(fields))
	res := make(map[string]interface{})
	_id, err := primitive.ObjectIDFromHex(query)
	if err != nil {
		log.Println("_id error", err)
		return &res, true
	}
	sr := m.C.Database(m.DbName).Collection(c).FindOne(m.Ctx, map[string]interface{}{"_id": _id}, of)
	if sr.Err() == nil {
		sr.Decode(&res)
	}
	return &res, true
}

//底层查询方法
func (m *MongodbSim) Find(c string, query interface{}, order interface{}, fields interface{}, single bool, start int, limit int) (*[]map[string]interface{}, bool) {
	defer catch()
	m.Open()
	defer m.Close()
	var res []map[string]interface{}
	coll := m.C.Database(m.DbName).Collection(c)
	if single {
		res = make([]map[string]interface{}, 1)
		of := options.FindOne()
		of.SetProjection(ObjToOth(fields))
		of.SetSort(ObjToM(order))
		if sr := coll.FindOne(m.Ctx, ObjToM(query), of); sr.Err() == nil {
			sr.Decode(&res[0])
		}
	} else {
		res = []map[string]interface{}{}
		of := options.Find()
		of.SetProjection(ObjToOth(fields))
		of.SetSort(ObjToM(order))
		if start > -1 {
			of.SetSkip(int64(start))
			of.SetLimit(int64(limit))
		}
		cur, err := coll.Find(m.Ctx, ObjToM(query), of)
		if err == nil && cur.Err() == nil {
			cur.All(m.Ctx, &res)
		}
	}
	return &res, true
}

func ObjToOth(query interface{}) *bson.M {
	return ObjToMQ(query, false)
}
func ObjToM(query interface{}) *bson.M {
	return ObjToMQ(query, true)
}

//obj(string,M)转M,查询用到
func ObjToMQ(query interface{}, isQuery bool) *bson.M {
	data := make(bson.M)
	defer catch()
	if s2, ok2 := query.(*map[string]interface{}); ok2 {
		data = bson.M(*s2)
	} else if s3, ok3 := query.(*bson.M); ok3 {
		return s3
	} else if s3, ok3 := query.(*primitive.M); ok3 {
		return s3
	} else if s, ok := query.(string); ok {
		json.Unmarshal([]byte(strings.Replace(s, "'", "\"", -1)), &data)
		if ss, oks := data["_id"]; oks && isQuery {
			switch ss.(type) {
			case string:
				data["_id"], _ = primitive.ObjectIDFromHex(ss.(string))
			case map[string]interface{}:
				tmp := ss.(map[string]interface{})
				for k, v := range tmp {
					tmp[k], _ = primitive.ObjectIDFromHex(v.(string))
				}
				data["_id"] = tmp
			}
		}
	} else if s1, ok1 := query.(map[string]interface{}); ok1 {
		data = s1
	} else if s4, ok4 := query.(bson.M); ok4 {
		data = s4
	} else if s4, ok4 := query.(primitive.M); ok4 {
		data = s4
	} else {
		data = nil
	}
	return &data
}
func intAllDef(num interface{}, defaultNum int) int {
	if i, ok := num.(int); ok {
		return int(i)
	} else if i0, ok0 := num.(int32); ok0 {
		return int(i0)
	} else if i1, ok1 := num.(float64); ok1 {
		return int(i1)
	} else if i2, ok2 := num.(int64); ok2 {
		return int(i2)
	} else if i3, ok3 := num.(float32); ok3 {
		return int(i3)
	} else if i4, ok4 := num.(string); ok4 {
		in, _ := strconv.Atoi(i4)
		return int(in)
	} else if i5, ok5 := num.(int16); ok5 {
		return int(i5)
	} else if i6, ok6 := num.(int8); ok6 {
		return int(i6)
	} else if i7, ok7 := num.(*big.Int); ok7 {
		in, _ := strconv.Atoi(fmt.Sprint(i7))
		return int(in)
	} else if i8, ok8 := num.(*big.Float); ok8 {
		in, _ := strconv.Atoi(fmt.Sprint(i8))
		return int(in)
	} else {
		return defaultNum
	}
}

//出错拦截
func catch() {
	if r := recover(); r != nil {
		log.Println(r)
		for skip := 0; ; skip++ {
			_, file, line, ok := runtime.Caller(skip)
			if !ok {
				break
			}
			go log.Printf("%v,%v\n", file, line)
		}
	}
}

//根据bsonID转string
func BsonIdToSId(uid interface{}) string {
	if uid == nil {
		return ""
	} else if u, ok := uid.(string); ok {
		return u
	} else if u, ok := uid.(primitive.ObjectID); ok {
		return u.Hex()
	} else {
		return ""
	}
}

func StringTOBsonId(id string) (bid primitive.ObjectID) {
	defer catch()
	if id != "" {
		bid, _ = primitive.ObjectIDFromHex(id)
	}
	return
}

func ToObjectIds(ids []string) []primitive.ObjectID {
	_ids := []primitive.ObjectID{}
	for _, v := range ids {
		_id, _ := primitive.ObjectIDFromHex(v)
		_ids = append(_ids, _id)
	}
	return _ids
}

//自动添加更新时间
func autoUpdateTime(db, coll string, ue *bson.M) {
	if db == "qfw" && coll == "user" {
		set := ObjToM((*ue)["$set"])
		if *set == nil {
			set = &bson.M{}
		}
		(*set)["auto_updatetime"] = time.Now().Unix()
		(*ue)["$set"] = set
	}
}

func IsObjectIdHex(hex string) bool {
	_, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return false
	}
	return true
}
