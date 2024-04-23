# gocrawler
gocrawler是非常轻量级的分布式爬虫框架， 可以快速构建高性能爬虫， 同时gocrawler严格遵循面向接口的设计， 所以gocrawler的各种组件都是可以轻松扩展的

更详细的说明， 可以参考[这里](https://superjcd.github.io/p/golang%E5%88%86%E5%B8%83%E5%BC%8F%E7%88%AC%E8%99%AB%E8%AE%BE%E8%AE%A1/)

## quick start
使用gocrawler的builder模式快速构建一个分布式爬虫, 作为一个示例， 我们将使用这个爬虫来抓取[zyte](https://www.zyte.com/blog/)上的博客信息  
在运行下示例前， 你需要确保已经安装并能够链接以下依赖：
- [nsq](https://nsq.io/)
- [mongodb](https://www.mongodb.com/)

### 安装gocrawler
```shell
go get https://github.com/superjcd/gocrawler
```

