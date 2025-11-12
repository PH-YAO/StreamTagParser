# StreamTagParser
一个用于流式解析XML标签的Go语言库，专门处理类似COT（Chain of Thought）标签的流式解析，支持自定义标签格式，支持中英等多语种文本。

## 功能特性
- ✅ 流式XML标签解析
- ✅ 标签识别和处理
- ✅ 状态机驱动的解析引擎
- ✅ 支持不完整标签的缓存和续传
- ✅ 可配置的标签格式
## 下载可用，无需安装
```
git clone https://github.com/PH-YAO/StreamTagParser.git
```
go run main.go stream_tag_parser.go

```
## 示例输出
***   streaming tag parser input: Opening Ceremony<tag=Watch with a happy mood>A classic play is being performed.</tag>Closing Ceremony   ***
previous text without TAG: Opening Ceremony
text: A classic play is being performed., tag: Watch with a happy mood
last text out of tag: Closing Ceremony

***   streaming tag parser input: 开幕式<tag=今天心情不错>这是一段话剧</tag>闭幕式   ***
previous text without TAG: 开幕式
text: 这是一段话剧, tag: 今天心情不错
last text out of tag: 闭幕式

***   streaming tag parser input: Opening Ceremony开幕式<tag=Watch with a happy mood今天心情不错>A classic play is being performed.这是一段话剧</tag>Closing Ceremony闭幕式   ***
previous text without TAG: Opening Ceremony开幕式
text: A classic play is being performed.这是一段话剧, tag: Watch with a happy mood今天心情不错
last text out of tag: Closing Ceremony闭幕式

## 开发说明
项目使用标准的Go模块结构，包含两个主要文件：
- main.go : 示例使用代码
- stream_tag_parser.go : 核心解析器实现

## 许可证
MIT License