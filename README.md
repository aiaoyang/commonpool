该仓库改自 https://github.com/fatih/pool
使用方式：
在自己的函数中创建一个factory函数，将需要生成的连接放入该函数中,然后调用NewPool()函数获得连接池，下面以postgres数据库的连接为例

```
var (
    factory = func () (io.Closer, error){
        return sql.Open("postgres", fmt.Sprintf("user=%s host=%s dbname=%s sslmode=%s password=%s", user, host, dbname， sslmode, password))
    }
)
func main(){
    P, err := NewPool(1, 1, factory)
    if err != nil {
        log.Printf("create Pool failed,err is: %s\n", err.Error())
    }

    //从连接池中获取连接
    v, err := P.Get()
    if err != nil {
        log.Printf("get connection from pool failed, err is: %s\n", err.Error())
    }
    
    //使用断言来使用连接池内资源, 实际使用中请根据实际情况进行断言
    v.(*pool.Conn).Client.(*sql.DB).DoSomething()
    ...
    
    //将使用玩的连接放回连接池
    v.Close()

    //如果希望真实的关闭该连接，需要使用MarkUnusable()方法
    //v.(*pool.Conn).MarkUnusable()
}
```