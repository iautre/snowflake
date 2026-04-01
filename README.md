## 一个 Go 的雪花 ID

### 特性

- 主 API：`NextID() (int64, error)`，时钟回拨、配置错误等返回 `error`，不再使用 `log.Fatal` 结束进程。
- 兼容：`NextId() int64` 仍保留，出错时 **panic**（与旧版遇错即退出相比，可被 `recover`）。
- 多实例：通过环境变量 `SNOWFLAKE_WORKER_ID`、`SNOWFLAKE_DATACENTER_ID`（均为 0~31）区分机器，避免多进程发号冲突。
- 链式配置：`WithTwepoch` / `WithWorkerId` / `WithDatacenterId` 非法时写入内部错误，后续 `NextID` 返回该错误（可 `errors.Is` 判断 `ErrInvalidWorkerID` 等）。

### 引入

```bash
go get github.com/iautre/snowflake
```

### 直接使用

```go
import "github.com/iautre/snowflake"

id, err := snowflake.NextID()
if err != nil {
    // 处理时钟回拨等
}
// 兼容旧代码（出错 panic）：
id64 := snowflake.NextId()
s := snowflake.String() // 等价于 FormatInt(NextId(), 10)，出错时 panic
```

### 自定义纪元、worker、数据中心

```go
sf := snowflake.NewSnowflake().WithTwepoch(1655647200000).WithWorkerId(1).WithDatacenterId(1)
id, err := sf.NextID()
str, err := sf.NextString()
```

### 环境变量（全局默认实例 `snowflake.ID`）

部署多副本时建议设置：

- `SNOWFLAKE_WORKER_ID`：0~31
- `SNOWFLAKE_DATACENTER_ID`：0~31

### 感谢

参考 [hutool](https://gitee.com/dromara/hutool) [Java Snowflake](https://gitee.com/dromara/hutool/blob/v5-master/hutool-core/src/main/java/cn/hutool/core/lang/Snowflake.java)
