## gopdf

![GitHub](https://img.shields.io/github/license/tiechui1994/gopdf)
![GitHub issues](https://img.shields.io/github/issues/tiechui1994/gopdf)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/tiechui1994/gopdf)

## 项目介绍

gopdf 是一个生成 `PDF` 文档的 `Golang` 库. 主要有以下的特点:

- 支持 Unicode 字符 (包括中文, 日语, 朝鲜语, 等等.)
- 文档内容的自动定位与分页, 减少用户的工作量.
- 支持图片插入, 格式可以是 `PNG` 或者 `JPEG`, 图片会进行适当压缩.
- 文档压缩
- 复杂表格组件, 块文本等

## 安装

```
go get -u github.com/tiechui1994/gopdf
```

## 案例展示: 

![image](./example/pictures/example.png)

代码参考 `example/complex_report_test`

![image](./example/pictures/table.png)

代码参考 `example/simple_table_test`

![image](./example/pictures/mutil-table.png)

代码参考 `example/mutil_table_test`


## 未来开发计划

1. 准备尝试开发 `Markdown` 的语法解析库, 然后通过解析库将 `markdown` 转换成 pdf, 可以支持定义一些颜色风格. 目前正
在研究 `marked.js` 前段库, 寻找灵感.

2. 开发更加通俗易用的组件, 比如 `paragraph`, `tablecell` 等.

