# gocrawler
gocrawler是非常轻量级的分布式爬虫框架， 可以快速构建高性能爬虫（生产者-消费者模式）， 同时gocrawler严格遵循面向接口的设计， 所以gocrawler的各种组件都是可以轻松扩展的

<!-- TOC -->

- [gocrawler](#gocrawler)
    - [快速开始](#%E5%BF%AB%E9%80%9F%E5%BC%80%E5%A7%8B)
        - [基础实现](#%E5%9F%BA%E7%A1%80%E5%AE%9E%E7%8E%B0)
        - [解析并提交更多Request](#%E8%A7%A3%E6%9E%90%E5%B9%B6%E6%8F%90%E4%BA%A4%E6%9B%B4%E5%A4%9Arequest)
        - [发送Request到其他爬虫Worker](#%E5%8F%91%E9%80%81request%E5%88%B0%E5%85%B6%E4%BB%96%E7%88%AC%E8%99%ABworker)
            - [第一步:替换默认Scheduler](#%E7%AC%AC%E4%B8%80%E6%AD%A5%E6%9B%BF%E6%8D%A2%E9%BB%98%E8%AE%A4scheduler)
            - [第二步：发送Request到seconndScheduler](#%E7%AC%AC%E4%BA%8C%E6%AD%A5%E5%8F%91%E9%80%81request%E5%88%B0seconndscheduler)
    - [自定义组件](#%E8%87%AA%E5%AE%9A%E4%B9%89%E7%BB%84%E4%BB%B6)
        - [替换网络请求组件](#%E6%9B%BF%E6%8D%A2%E7%BD%91%E7%BB%9C%E8%AF%B7%E6%B1%82%E7%BB%84%E4%BB%B6)
            - [追加请求头](#%E8%BF%BD%E5%8A%A0%E8%AF%B7%E6%B1%82%E5%A4%B4)
        - [替换存储组件](#%E6%9B%BF%E6%8D%A2%E5%AD%98%E5%82%A8%E7%BB%84%E4%BB%B6)
        - [其他组件](#%E5%85%B6%E4%BB%96%E7%BB%84%E4%BB%B6)
    - [请求去重](#%E8%AF%B7%E6%B1%82%E5%8E%BB%E9%87%8D)
    - [任务计数](#%E4%BB%BB%E5%8A%A1%E8%AE%A1%E6%95%B0)
    - [错误处理和生命周期函数](#%E9%94%99%E8%AF%AF%E5%A4%84%E7%90%86%E5%92%8C%E7%94%9F%E5%91%BD%E5%91%A8%E6%9C%9F%E5%87%BD%E6%95%B0)
        - [沟通时机](#%E6%B2%9F%E9%80%9A%E6%97%B6%E6%9C%BA)
        - [沟通方式](#%E6%B2%9F%E9%80%9A%E6%96%B9%E5%BC%8F)
    - [Dev模式](#dev%E6%A8%A1%E5%BC%8F)
    - [参考](#%E5%8F%82%E8%80%83)

<!-- /TOC -->
## 快速开始
### 基础实现
使用gocrawler的builder模式能够快速构建一个分布式爬虫, 作为一个示例， 我们将使用gocrawler抓取[zyte](https://www.zyte.com/blog/)上的博客信息  
在运行下示例前， 你需要确保已经安装并能够链接以下依赖：
- [nsq](https://nsq.io/)
- [mongodb](https://www.mongodb.com/)

> gocrawler本身并不依赖nsq作为消息组件， 同样也不依赖mongodb作为存储组件，后面会介绍替换的方式

我们的目标是爬取[zyte网站](https://www.zyte.com/blog)上的所有blog的基础信息， 包括：
- 标题
- 作者
- 阅读时间
- 发布时间

> 我们会抓取列表项信息， 至于如何同时在抓取列表信息的同时抓取每个列表项的详情信息后面会介绍

首先我们创建一个项目
```
mkdir zyte
```
然后初始化项目
```
go mod init zyte
```
首先我们在zyte目录下创建一个parser目录， 并编写我们的解析函数：
```go
//parser/parser.go
package parser

import (
	"context"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/superjcd/gocrawler/parser"
)

type zyteParser struct{}

func NewZyteParser() *zyteParser {
	return &zyteParser{}
}

func (p *zyteParser) Parse(ctx context.Context, r *http.Response) (*parser.ParseResult, error) {
	doc, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		return nil, err
	}
	result := &parser.ParseResult{}
	resultItems := make([]parser.ParseItem, 0)

	doc.Find("div.CardResource_card__BhCok").Each(
		func(i int, s *goquery.Selection) {
			item := parser.ParseItem{}
			item["title"] = s.Find("div.free-text").Text()
			item["author"] = s.Find("div:nth-child(3) > div:nth-child(1) > span:nth-child(2)").Text()
			item["read_time"] = s.Find("div:nth-child(3) > div:nth-child(2) > span:nth-child(2)").Text()
			item["post_time"] = s.Find("div:nth-child(4) > div:nth-child(1) > span:nth-child(2)").Text()
			resultItems = append(resultItems, item)
		},
	)
	result.Items = resultItems

	return result, nil
}

```
> 推荐使用[goquery](https://github.com/PuerkitoBio/goquery)来构建网页解析组件

接着， 我们可以在`main.go`文件中正式构建我们的第一个爬虫:
```go
// main.go
package main

import (
	default_builder "github.com/superjcd/gocrawler/builder/default"
	"github.com/superjcd/gocrawler_examples/zyte/parser"
)

func main() {
	config := default_builder.DefaultWorkerBuilderConfig{}
	worker := config.Name("zyte").MaxRunTime(300).Workers(10).LimitRate(10).Build(parser.NewZyteParser())
	worker.Run()
}
```
在main.go路径运行命令`go run .`就能顺利地启动爬虫。当然为了让我们的爬虫worker工作起来， 我们需要喂给worker一些任务；
在pub/main.go中编写提交任务的逻辑(生产者)：
```go
package main

import (
	"fmt"
	"log"

	"github.com/gofrs/uuid"
	"github.com/superjcd/gocrawler/request"
	"github.com/superjcd/gocrawler/scheduler"
	"github.com/superjcd/gocrawler/scheduler/nsq"
)

func main() {
	s := nsq.NewNsqScheduler("zyte", "default", "127.0.0.1:4150", "127.0.0.1:4161")
	pages := []int{}
	for i := 1; i < 10; i++ {
		pages = append(pages, i)
	}
	uid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	log.Printf("taskId: %s", uid.String())

	for _, pg := range pages {
		data := make(map[string]string, 0)
		data["taskId"] = uid.String()
		url := fmt.Sprintf("https://www.zyte.com/blog/page/%d", pg)
		fmt.Println(url)
		req := request.Request{
			URL:    url,
			Method: "GET",
			Data:   data,
		}
		s.Push(scheduler.TYP_PUSH_SCHEDULER, &req)

	}
}
```
新开一个终端， 并运行`go run .\pub\`， 可以在启动woker的终端中看到目标网站被解析并存入到mongodb的日志信息。  
检查本地的mongodb的zyte数据库的default集合，你就会看到我们抓到的列表数据。


### 解析并提交更多Request
上面的例子有一个很大的问题在于：生产者显式地把需要抓取的page一页一页地提交给了gocrawler, 比如在上面例子中, 我们提交了9个请求， 问题是在真实场景下， 任务的请求数有可能是不固定的， 理想情况下， 我们会希望爬虫能够在爬取第一页的时候， 通过解析首页的最大页码数来自动的提交更多请求。    
这一点在gocrawler中很好实现，因为gocrawler的Parser组件的Parse函数产出的`*parser.ParseResult`的结构体是可以包含Request对象的， 而这些被解析出来的Request对象会被gocrawler提交
> 当然这里会衍生出另外的问题是， 如何过滤重复请求以及如何使用类似于自动的URL匹配器获取目标url， 关于前者， gocrawler可以通过添加Visit组件来过滤一定时间内已经抓取过的url， 后者gocrawler自身没有实现， 但是这个功能用户可以在自定义的Parser组件中实现

废话不多说 ，我们切入正题：
首先我们需要修改一下Parser:
```go
package parser

import (
	"context"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/superjcd/gocrawler/parser"
	"github.com/superjcd/gocrawler/request"
)

type zyteParser struct{}

func NewZyteParser() *zyteParser {
	return &zyteParser{}
}

func (p *zyteParser) Parse(ctx context.Context, r *http.Response) (*parser.ParseResult, error) {
    ...
    resultItems := make([]parser.ParseItem, 0)
	requests := []*request.Request{}
	// gocrawler会默认把Request对象中的Data属性传递到上下文中， 用户可以通过ctx.Value(request.RequestDataCtxKey{})来获取这个值（map）  
	ctxValue := ctx.Value(request.RequestDataCtxKey{})
	requestData := ctxValue.(map[string]string)
	page := requestData["page"]

	if page == "1" {
		uid, _ := uuid.NewV4()
		for pg := 2; pg <= 5; pg++ {
			data := make(map[string]string, 0)
			data["taskId"] = uid.String()
			data["page"] = strconv.Itoa(pg)
			url := fmt.Sprintf("https://www.zyte.com/blog/page/%d", pg)
			// 注意： 在这里我们构建新的请求
			req := request.Request{
				URL:    url,
				Method: "GET",
				Data:   data,
			}
			requests = append(requests, &req)
		}
	}
   
    ...
	result.Items = resultItems
	result.Requests = requests

	return result, nil
}
```
这样,当我们请求第一页的时候，就可以连带把其他页面的请求一并传递给任务队列(当然正常情况下, 最大页码数这个值是需要自己去解析的）


### 发送Request到其他爬虫Worker
如果我们想要把请求传递给其他的woker该怎么办呢， 假设我们有两个爬虫worker：
- 列表worker, 获取列表项
- 详情worker, 获取每一页的详情信息

这种需要用到多个worker的场景其实非常常见， 比如以抓取房价信息为例， 房屋的简要信息会以列表页形式存在， 比如一个列表页上面可能有20个房屋链接；然后当我们点击每个链接， 就可以获得该房屋的详情信息；    
由于列表页和详情页的url以及页面信息通常是不同的， 所以比较合理的方式就是分别运行两个Worker(可以共用部分组件， 比如fetcher), 那么现在需要面对的问题是， 如何在**列表爬虫**抓取列表页信息的时候， 把详情页的请求提交到**详情爬虫**？  
在gocrawler中实现这个方式只需要两步：
#### 第一步:替换默认Scheduler
gocrawler中的Scheduler组件有一个Option（选项）是secondScheduler（也是一个Scheduler接口）， 如果secondScheduler非空， 那么我们就能把请求传递给这个seconndScheduler（如何传递请求， 第二个步骤会讲）, 只要另外一个爬虫Worker订阅了seconndScheduler的消息，那么第二个worker自然也能同时进行运行。  

```go
package main 

import (
	"github.com/superjcd/gocrawler/worker"
	"github.com/superjcd/gocrawler/scheduler/nsq"
)


func main() {
	// 重新准备一个scheduler
	secondScheduler := nsq.NewNsqScheduler(your_second_topic, your_second_channel, "localhost:4150", "localhost:4161")
	scheduler := nsq.NewNsqScheduler(your_second_topic, your_second_channel, "localhost:4150", "localhost:4161", nsq.WithSecondScheduler(secondScheduler))

	config := default_builder.DefaultWorkerBuilderConfig{}
	worker := config.Name("zyte").MaxRunTime(300).Workers(10).LimitRate(10).Build(
		parser.NewZyteParser(),worker.WithScheduler(scheduler),) // 替换掉默认shceduler
	worker.Run()
}

```
> `Scheduler`是构建gocrawler引擎的一个重要组件，而在gocrawler中所有的组件都是接口，所以用户可以轻松进行替换；其他组件的替换可以详见下面的[自定义组件](#自定义组件)
 
#### 第二步：发送Request到seconndScheduler
要想把Request发送到secondScheduler很简单，只要修改一下Request的IsSecondary字段就好， 将它设置为true就可以了， 例如:  
假设我们在列表页抓到若干个详情页的url, 我们需要像上例一样在Parse函数中构造新的Request对象
```go
...    
	for _, url := range urls{ // urls是详情页请求地址队列
		reqData := make(map[string]string, 0)
		reqData["taskId"] = uid.String()
		newRequest := request.Request{
			URL:         url,
			Method:      "GET",
			Data:        reqData,
			IsSecondary: true,   // 这里是关键
		}
		requests = append(requests, &newRequest)
	}
    ...
	result.Items = resultItems
	result.Requests = requests

```

## 自定义组件
### 替换网络请求组件
gocrawler的默认Fetcher只是一个非常简单的网络请求组件，只使用默认网络请求组件在应对一些常见的反扒手段的时候肯定是远远不够的， 所以我们有时候我们希望Fetcher可以支持
- 从代理池获取代理
- 从Cookie池获取cookie
- 改变请求头

Fetcher在gocrawler中只是一个接口，接口定义如下: 

```go
// Fetcher的定义
type Fetcher interface {
	Fetch(ctx context.Context, req *request.Request) (*http.Response, error)
}
```
如果想要替换掉默认Fetcher, 只要在在Build函数中添加`worker.WithFetcher(your_fetcher)`即可：
```go
config := default_builder.DefaultWorkerBuilderConfig{}
worker := config.Name("zyte").Build(your_parser, worker.WithFetcher(your_storage))
```

当然我会更推荐使用gocrawler中的NewFetcher去创建一个Fetcher对象, 结和Option模式， 将你需要的替换的部分(比如下面的代理获取组件)替换掉即可：

```go
import (
	"time"
	"github.com/superjcd/gocrawler/fetcher" 
)

fetcher := fetcher.NewFetcher(10 * time.Second, fetcher.WithProxyGetter(your_proxy_getter))
```
`your_proxy_getter`是你需要实现的proxy获取组件, 它的定义如下：
```
type ProxyGetter interface {
	Get(*http.Request) (*url.URL, error)
}
```
所以， 如果你需要从你的代理池中获取你的代理， 然后通过代理发起请求， 你只要去自己去实现`ProxyGetter`即可

#### 追加请求头
请求头是默认的Fetcher组件的一部分，如果用户想要添加请求头， 可以通过下面的方式进行实现：
```go
import (
	"time"
	"github.com/superjcd/gocrawler/fetcher" 
)

headers := map[string]string{
	"accept": "application/json"
}

fetcher := fetcher.NewFetcher(10 * time.Second, fetcher.WithHeaders(headers))
```
`User-Agent`也是请求头的一部分, 用户可以基于上面的方式进行添加， 或者使用`UaGetter`动态地设置User-Agent，例如:
```go
import (
	"time"
	"github.com/superjcd/gocrawler/fetcher" 
	"github.com/superjcd/gocrawler/ua" 
)

uaGetter :=  ua.NewRoundRobinUAGetter()
fetcher := fetcher.NewFetcher(10 * time.Second, fetcher. WithUaGetter(uaGetter))
```
> uaGetter会在每一次Fetcher进行网络请求的时候， 从一个随机UA池中挑选一个user-agent;在默认的Build模式中, 默认fetcher会自定使用这个特性


### 替换存储组件
gocrawler的`DefaultWorkerBuilderConfig`目前只支持使用mongodb来作为爬虫的默认存储组件， 如果用户想要使用别的存储组件， 只要实现一个自定义的Storage即可，然后和前面的自定义Fetcher类似， 通过在Build函数中添加`worker.WithStorage(your_storage)`就能替换掉默认存储组件:
```go
type Storage interface {
	Save(ctx context.Context, datas ...parser.ParseItem) error
}
```
需要注意的是， 用户自定义存储组件的时候， 最好考虑结合一些缓存机制，比如当缓存收集到一定数量的对象之后再把数据flush到存储器， 而不是一条一条数据的存， 特别是对于mysql这类关系数据库而言，高并发下使用逐条存储的代价是很大的。  
默认的mongo存储组件是考虑了缓存机制的，用户可以通过调用`DefaultWorkerBuilderConfig`的`BufferSize`和`AutoFlushInterval`来定义缓存大小以及flush间隔（秒）， 例如：
```go
config := default_builder.DefaultWorkerBuilderConfig{}
worker := config.Name("zyte").Workers(10).LimitRate(10).BufferSize(100).AutoFlushInterval(10).Build(your_parser, worker.WithStorage(your_storage))
```
在上例中， 我们的爬虫有一个大小为100的缓存，缓存如果满了就会存储到mongo中， 如果缓存没有满，也会在10秒之后被flush到mongo中

### 其他组件
- [Visit](https://github.com/superjcd/gocrawler/blob/main/visit/visit.go) 去重组件
- [Counter](https://github.com/superjcd/gocrawler/blob/main/counter/counter.go) 任务计数组件

这些组件都可以通过`Build(parser, With<组件>(组件实现))`来嵌入到gocrawler中，或者说替换掉默认组件

## 请求去重
我们希望在爬虫的某个运行周期中， 不想重复请求， 可以使用Visit组件进行去重
Visit组件的接口定义如下：
```go
type Visit interface {
	SetVisitted(key string, ttl time.Duration) error
	UnsetVisitted(key string) error
	IsVisited(key string) bool
}
```
`SetVisitted`会将某个请求在一定的声明周期内(ttl)会被标记为已被访问， 被标记过的请求(也就是Request对象)不会在这个周期内被再次访问
gocrawler中有可以通过一下方式，通过redis来实现请求去重:
```go
package main 

import (
	"github.com/superjcd/gocrawler/worker"
	"github.com/superjcd/gocrawler/vist/redis"
)
config := default_builder.DefaultWorkerBuilderConfig{}
worker := config.Name("zyte").Build(your_parser, worker.WithVisiter(redis.NewRedisVisit(redis.Options, prefixKey)))
```
> gocrawler会默认根据Request对象的Url和Method进行去重，如果想要添加`Request.Data`中值作为去重项，通过在Build函数中使用`worker.WithAddtionalHashKeys(your_keys)`来实现， 注意如果你指定的key不存在于`Request.Data`，会panic
## 任务计数
对分布式爬虫进行任务计数会有一些麻烦，目前gocraler默认提供的`redisTaskCounters`基于redis的乐观锁机制实现了一个可用的分布式计数， 使用方式和上诉其他组件类似，不再赘述 

## 错误处理和生命周期函数
由于爬虫需要和网络以及各种日新月异的反爬技术打交道， 所以关于爬虫任务， 有一点是不会错的：
> 我们的爬虫随时都会出错

所以如何正确的处理错误的请求是爬虫任务的一个挥之不去的主题， 简单的丢弃失败的请求肯定是不可行的， 当然无限的重试自然也不可取， 有限次数的重试似乎是不错的折衷方法， gocrawler也是这么做的， 重试次数用户可以通过`DefaultWorkerBuilderConfig`的`Retries`方法来定义(默认是5次) ，但是还有一个更加关键的点在于--用户如何告诉gocrawler对某个失败的请求进行重试而不是丢弃呢， 因为有时候我们确实也需要丢弃掉不需要的请求(比如状态码是404的请求), 所以这种和gocrawler引擎进行沟通的机制是必要。  
实现这个沟通机制的关键在于两点：
- 沟通的时机
- 沟通的方式

### 沟通时机
gocrawler的Worker有以下生命周期函数：
- BeforeRequest 发生在请求之前
- AfterRequest  发生在请求之后
- BeforeSave    发生在存储之前
- AfterSave     发生在存储之后

这里以`AfterRequest`为例：
```go
func (w *worker) AfterRequest(ctx context.Context, resp *http.Response) (Signal, error) {
	var sig Signal
	if w.AfterRequestHook != nil {
		return w.AfterRequestHook(ctx, resp)
	}
	sig |= DummySignal
	return sig, nil
}
```
`AfterRequest`会发生在请求发生之后（Fetcher进行fetch之后）， 页面被解析之前；如果用户提供了`AfterRequestHook`，那么`AfterRequestHook`就会在这个阶段被执行（一个生命周期函数会对应一种hook）；  
所以用户完全可以在这个阶段， 通过判断请求的状态码来确定是不是要进行重试

### 沟通方式
说完了沟通时机， 现在需要说一下方式了；gocrawler会基于生命周期函数返回的Signal来决定下一步该如何行动， 下面我们尝试定义一个 `AfterRequestHook`（它会返回Signal）：
```golang
import "github.com/superjcd/gocrawler/worker"
func CheckResponseStatus(ctx context.Context, resp *http.Response) (worker.Signal, error) {
	var sig worker.Signal
	switch resp.StatusCode {
	case http.StatusOK:
		sig |= worker.DummySignal
	case http.StatusNotFound:
		sig |= worker.ContinueWithoutRetrySignal
	default:
		sig |= worker.ContinueWithRetrySignal
	}

	return sig, nil
}
```
> 用户可以通过在Build函数中添加worker.WithAfterRequestHook(CheckResponseStatus)来注册这个hook,其他生命周期的hook的注册方式也是一样的
`CheckResponseStatus`会去判断http.Response的状态码， 如果是200就返回`DummySignal`信号， 404就返回`ContinueWithoutRetrySignal`,在其他情况下就是返回`ContinueWithRetrySignal`信号；  
当gocrawler接收到`DummySignal`的， 会继续执行； 接收到`ContinueWithoutRetrySignal`的时候则会跳过后面的步骤直接处理下一个请求；而接收到`ContinueWithRetrySignal`的时候， gocrawler就会发起重试， 完整的信号列表：
```golang
type Signal int8

const (
	DummySignal = 1 << iota    //  默认初始signal
	ContinueWithRetrySignal    //  重试信号
	ContinueWithoutRetrySignal  // 不重试, 继续下一个任务
	BreakWithPanicSignal        //  停止爬虫并panic
	BreakWithoutPanicSignal     //  停止爬虫但是不panic
)
```
> Signal本质上就是一个8位有符号整数

（最后还有一点需要注意的是， 上面的重试并不是立马重试， 而是请求会被重新发送到请求队列中，等待下一次被处理）



## Dev模式
TODO

## 参考
- [我的博客](https://superjcd.github.io/p/golang%E5%88%86%E5%B8%83%E5%BC%8F%E7%88%AC%E8%99%AB%E8%AE%BE%E8%AE%A1/)
- [文档中的zyte的例子](https://github.com/superjcd/gocrawler_examples)