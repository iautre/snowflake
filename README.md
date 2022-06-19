## 一个 go 的雪花id

### 使用
#### 引入包
```
go get github.com/iautre/snowflake

```
#### 直接使用
```
import (
	……
	"github.com/iautre/snowflake"
    ……
)
func Test(){
    var id int64 = snowflake.NextId()
    var idStr string = snowflake.String()
}
```
#### 自定义开始时间，工作机器节点id，数据中心ID
```
import (
	……
	"github.com/iautre/snowflake"
    ……
)
func Test(){
    sf := snowflake.NewSnowflake().WithTwepoch(1655647200000).WithWorkerId(1).WithDatacenterId(1)
    var id int64 = sf.NextId()
    var idStr string = sf.String()
}