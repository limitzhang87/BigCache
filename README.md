# BigCache

> 来源 https://dev.to/calebschoepp/storing-empty-interfaces-in-bigcache-1b33

最近工作中，我的任务是向我们的一个 Goland 服务添加缓存。这个服务需要传入请求以提供身份身份认证的  API key。因此，对于每个请求，该服务都会额外查询数据库以验证 API key，尽管它通常是相同的 key。这很不好，实现缓存最终比我想象的要难很多。

经过调研和工程师之间详尽讨论之后，我们认为 **BigCache**最适合我们的需求。

这里有一个问题，BigCache 中的 set 方法的声明为 ``Set(key string, entry []byte) error``。他期望存储一个``byte slice``，但是我们希望存储一个 ``struct``，该 struct 具有多个 API key 的字段。这次我们可能只能够存储实际 key 的 bytes。但这只是推迟解决问题。我们希望的是像其他 Goland 缓存实现中的声明 ``Set(key string , entry interface{})``。这样我们就可以存储任何我们想要的东西。

这个问题的明显解决方案是序列化，如果我们可以将任意结构序列化为 byte slice，那么我们就可以存储任意内容。要使用我们存储的结构，可以从缓存中反序列化获取到的 byte slice。序列化结构就像在 Goland 中导入任意数量的可用 encoding 库一样容易，但是现在头疼的问题来了。当我们反序列化 bytes 时，Go 语言如何知道将数据存入什么类型的结构？事实证明，Goland 特有的反序列化库 ``encoding/gob`` 具有此功能。

我强烈建议阅读 Rob Pike 写的关于 Gob 的博客文章，这是一篇好文章。简而言之，Gob 是一种 Go 原生的数据库系列反序列化方式，它还具有序列化 interface 类型的功能。为此，您需要在序列化之前使用恰当命名的 **register function**注册您的类型。我在这里卡出了，因为我找到了关于``register()``的任何代码示例总是要注册一个单一的 struct 或 interface；我需要注册任意 ``interface{}``类型。我在 Go playground 上摸索了一下发现它也可以做到。

```go
// 大多数示例中注册的类型
type foo struct {
    bar string
}

gob.register(foo{})

// 我想要注册的类型
var type interface{}

gob.register(type)
```

#### 把它们组合在一起

解决了将任意 struct 存储为 bytes 的问题后，下面将展示如何将他们组合到一起，首先，需要一个缓存 interface， 以便系统的其余部分能够与之交互。对于一个简单的缓存，我们只需要 get 和 set 方法。

```go
type Cache struct {
    Set(key, value interface{}) error
    Get(key interface{}) (interface{}, error)
}
```

现在，让我们实现上述接口的 BigCahe 实现。首先，我们需要一个结构体来保存缓存并可以向其中添加方法。还可以在此结构中添加其他字段，例如 metrics。

```go
type bigCache struct {
    cache *bigcache.BigCache
}
```

接下来是 get 和 set 方法的实现。两种方法都断言 key 是 string。由此开始，get 和 set 的实现都彼此独立了。一个序列化一个值并存储它。另一个获取值并将其反序列化。

```go
func (c *bigCache) Set(key, value interface{}) error {
    // 断言 key 为 string 类型
    keyString, ok := key.(string)
    if !ok {
        return errors.New("a cache key must be a string")
    }
    // 将 value 序列化为 bytes
    valueBytes, err := serialize(value)
    if err != nil {
        return err
    }
    return c.cache.Set(keyString, valuesBytes)
}

func (c *bigCache) Get(key interface{}) (interface{}, error) {
    // 断言 key 为 string 类型
    keyString, ok := key.(string)
    if (!ok) {
        return nil, errors.New("a cache key must be a string")
    }
    
    // 获取以bytes 格式存储的 value
    valueBytes, err := c.cache.Get(keyString)
    if err != nil {
        return nil, err
    }
    
    // 反序列化
    value, err := deserialize(valueBytes)
    if err != nil {
        return nil, err
    }
    return value, nil
}
```

最后是 ``encoing/gob``序列化逻辑。除了使用 ``register()``之外，这是 Go 中序列化内容相当标准的用法。

```go
func serialize(value interface{}) ([]byte, error) {
    buf := bytes.Buffer{}
    enc := gob.NewEncoder(&buf)
    gob.Register(value)
    
    err := enc.Encode(&value)
    if err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

func deserialize(valueBytes []byte) (interface{}, error) {
    var value interface{}
    buf := bytes.NewBuffer(valueBytes)
    dec := gob.NewDecoder(buf)
    
    err := dec.Decode(&value)
    if err != nil {
        return nil, rrr
    }
    return value, nil
}
```

通过这些，我们已经设法在 BigCache 中存储 ``interface{}`` 值了。



