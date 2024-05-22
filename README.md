# gocrawler
gocrawler是非常轻量级的分布式爬虫框架， 可以快速构建高性能爬虫（生产者-消费者模式）， 同时gocrawler严格遵循面向接口的设计， 所以gocrawler的各种组件都是可以轻松扩展的

更详细的说明， 可以参考[这里](https://superjcd.github.io/p/golang%E5%88%86%E5%B8%83%E5%BC%8F%E7%88%AC%E8%99%AB%E8%AE%BE%E8%AE%A1/)；文档中的例子，可以参考[这里](https://github.com/superjcd/gocrawler_examples)

## 快速开始
使用gocrawler的builder模式能够快速构建一个分布式爬虫, 作为一个示例， 我们将使用gocrawler抓取[zyte](https://www.zyte.com/blog/)上的博客信息  
在运行下示例前， 你需要确保已经安装并能够链接以下依赖：
- [nsq](https://nsq.io/)
- [mongodb](https://www.mongodb.com/)

我们的目标是爬取[zyte网站](https://www.zyte.com/blog)上的所有blog的基础信息， 包括：
- 标题
- 作者
- 阅读时间
- 发布时间

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

接着， 我们可以在main文件中正式构建我们的第一个爬虫:
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
运行`go run .`就能顺利地启动爬虫。当然为了验证爬虫worker是不是真正在运行， 我们需要喂给worker一些任务；
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
新开一个terminal并运行`go run .\pub\`， 可以在启动woker的终端中看到目标网站被解析并存入到mongodb的日志信息。  
检查本地的mongodb的zyte数据库的default集合，你就会看到你想要的数据。就是这么简单



## 解析并提交更多Request
上面的例子有一个很大的问题在于，生产者显式地把需要抓取的page一页一页地提交给了gocrawler, 比如在上面例子中, 我们提交了9个请求， 问题是在真实场景下， 任务的请求数有可能是不固定的， 理想情况下， 我们会希望爬虫能够解析并递交请求。  
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
> gocrawler会默认把Request对象中的Data属性传递到上下文中， 用户可以通过ctx.Value(request.RequestDataCtxKey{})来获取这个值  

这样,当我们请求第一页的时候， 我们通过首页得到的最大页码数(5)， 就可以连带把其他页面的请求一并传递给任务队列(当然正常情况下, 最大页码数这个值是需要自己去解析的）


## 发送Request到其他爬虫Worker
如果我们想要把请求传递给其他的woker该怎么办呢， 假设我们有两个爬虫worker：
- 列表worker, 获取列表项
- 详情worker, 获取每一页的详情信息

这种需要用到多个worker的场景其实非常常见， 比如以抓取房价信息为例， 房屋的简要信息会以列表页形式存在， 比如一个列表页上面可能有20个房屋链接；然后当我们点击每个链接， 就可以获得该房屋的详情信息；    
由于列表页和详情页的url以及页面信息通常是不同的， 所以比较合理的方式就是分别运行两个Worker(可以共用部分组件， 比如fetcher), 那么现在需要面对的问题是， 如何在**列表爬虫**抓取列表页信息的时候， 把详情页的请求提交到**详情爬虫**？  
在gocrawler中实现这个方式只需要两步：
### 第一步:替换默认Scheduler
gocrawler中的Scheduler组件有一个Option（选项）是secondScheduler（也是一个Scheduler接口）， 如果secondScheduler非空， 那么我们就能把请求传递给这个seconndScheduler（如何传递请求， 第二个步骤会讲）, 只要另外一个爬虫Worker订阅了seconndScheduler，那么第二个worker自然也能同时进行运行。

首先我们通过调用`DefaultWorkerBuilderConfig`的`NsqScheduler`,`NsqScheduler`会为我们的列表worker的Scheduler对象添加一个seconndScheduler， 然后用这个带seconndScheduler的Scheduler替换默认的Scheduler： 

```go
...(略)

func main() {
	config := default_builder.DefaultWorkerBuilderConfig{}
	worker := config.Name("zyte").MaxRunTime(300).Workers(10).LimitRate(10).NsqScheduler("", "","list_worker", "default", "details_worker", "default").Build(parser.NewZyteParser())
	worker.Run()
}

```
NsqScheduler接受6个参数，前两个是nsq的地址参数，可以像上面这样使用默认值(默认nsq按照官网会运行在本地)；后面四个分别是`topic`, `channel`,`second_topic`, `second_channel`， 前两个定义了主worker(也就是列表worker)的消息队列的topic和channel参数， 后面两个就是我们的seconndScheduler的topic和channel参数。

### 第二步：发送Request到seconndScheduler
要想把Request发送到secondScheduler很简单，只要修改一下Request的IsSecondary字段就好， 将它设置为true就可以了， 例如:  
假设我们在列表页抓到若干个详情页的url, 我们需要像上例一样在Parse函数中构造新的Request对象
```go
...    
	for _, url := range urls{ // urls是详情页请求地址队列
		reqData := make(map[string]string, 0))
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