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
