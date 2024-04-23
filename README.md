# gocrawler
gocrawler是非常轻量级的分布式爬虫框架， 可以快速构建高性能爬虫（生产者-消费者模式）， 同时gocrawler严格遵循面向接口的设计， 所以gocrawler的各种组件都是可以轻松扩展的

更详细的说明， 可以参考[这里](https://superjcd.github.io/p/golang%E5%88%86%E5%B8%83%E5%BC%8F%E7%88%AC%E8%99%AB%E8%AE%BE%E8%AE%A1/)

## 快速开始
使用gocrawler的builder模式能够快速构建一个分布式爬虫, 作为一个示例， 我们将使用gocrawler抓取[zyte](https://www.zyte.com/blog/)上的博客信息  
在运行下示例前， 你需要确保已经安装并能够链接以下依赖：
- [nsq](https://nsq.io/)
- [mongodb](https://www.mongodb.com/)

我们的目标是爬取网站https://www.zyte.com/blog上的所有blog的基础信息， 包括：
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
> 推荐使用gocrawler来构建网页解析组件

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


更多例子, to be continued...
