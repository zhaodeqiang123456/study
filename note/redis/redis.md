#### Redis 操作

##### 连接redis 服务

```go
import "github.com/redis/go-redis/v9"
var rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
ctx := context.Background()
rdb.Ping(ctx)
```

##### string 缓存

```go
rdb.Set(ctx, "task:task-1", `{"status":"pending"}`, 30*time.Second)
val, _ := rdb.Get(ctx, "task:task-1").Result()
```

##### 删除与过期

~~~go
rdb.Del(ctx, "task:task-1")
rdb.Expire(ctx, "task:task-1", 60*time.Second)
~~~

##### hash存储对象

~~~go
rdb.HSet(ctx, "user:1", "name", "张三", "age", 25)
name, _ := rdb.HGet(ctx, "user:1", "name").Result()
~~~


---



#### Redis 缓存穿透防护


> 应用场景： 生成批量的不存在的key, 进行查询，redis 一定失效，因此会直接穿透到数据库，对数据库造成极大压力
>
> 目的: 为了应对非法访问，缓解数据库压力

实现方式： 对于不存在的key, 访问数据库后，在redis中设定一定时长的NULL缓存，以防后续访问直接穿透到数据库

```go
func getTaskWithCache(taskID string) (*Task, error) {
    // 1. 查缓存
    cached, err := rdb.Get(ctx, "task:"+taskID).Result()
    if err == nil {
        if cached == "NULL" {
            return nil, ErrNotFound // 命中空值标记
        }
        var task Task
        json.Unmarshal([]byte(cached), &task)
        return &task, nil
    }

    // 2. 查数据库
    var task Task
    err = db.QueryRow("SELECT id, status, result FROM tasks WHERE id = ?", taskID).
        Scan(&task.ID, &task.Status, &task.Result)
    if err == sql.ErrNoRows {
        // 存空值，防止穿透，过期时间短一些
        rdb.Set(ctx, "task:"+taskID, "NULL", 30*time.Second)
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, err
    }

    // 3. 回写缓存（正常数据）
    taskJSON, _ := json.Marshal(task)
    rdb.Set(ctx, "task:"+taskID, taskJSON, 5*time.Minute)
    return &task, nil
}
```


#### Redis 缓存击穿防护

> 应用场景：当缓存失效时, 对某个特定的数据行，产生大量的请求，此时全部打到数据库，给数据库造成巨大压力
>
> 目的: 缓解数据库压力，同时预防可能的异步问题

实现方式：对数据库的访问加互斥锁，同一时刻只允许一个访问

```go
var (
    lockMap   = make(map[string]*sync.Mutex)
    lockMapMu sync.Mutex
)

func getOrCreateLock(key string) *sync.Mutex {
    lockMapMu.Lock()
    defer lockMapMu.Unlock()
    if l, ok := lockMap[key]; ok {
        return l
    }
    l := &sync.Mutex{}
    lockMap[key] = l
    return l
}

func getTaskWithMutex(taskID string) (*Task, error) {
    // 先查缓存
    cached, err := rdb.Get(ctx, "task:"+taskID).Result()
    if err == nil && cached != "NULL" {
        var task Task
        json.Unmarshal([]byte(cached), &task)
        return &task, nil
    }

    // 获取专属锁，防止击穿
    mu := getOrCreateLock("task-lock:" + taskID)
    mu.Lock()
    defer mu.Unlock()

    // 双重检查
    cached, err = rdb.Get(ctx, "task:"+taskID).Result()
    if err == nil && cached != "NULL" {
        var task Task
        json.Unmarshal([]byte(cached), &task)
        return &task, nil
    }

    // 查数据库并回写（逻辑同上）
    // ...
}
```
