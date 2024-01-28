# go-library

#### 介绍
在学习和工作中整理的一些常见的辅助类和函数，相当于我的 framework


#### 功能列表

1. debugutil: 一些为了方便开发和调试的辅助函数
   - verifyXxx 
     - 使用 CSTD(Code Self Test  Development) 技术方式辅助处理 error(不需要编写大量的 if 来处理error,又不会遗漏 error)
     - 参见文档: [Write bug-free code with CSTD(Code Self Test  Development)/使用CSTD技术轻松编写0Bug的代码](doc/使用CSTD技术轻松编写0Bug的代码.doc) 
2. flog: 为了在 verify 中使用,定义的简单日志封装接口,一般来说,用户需要使用 `SetLoggerFactory` 进行定制
3. mime/multipart/VirtualWriter:
   - 类似于内置的 `multipart.Writer`,但是能在不耗费大量内存的前提下同时上传很多(4G+)文件
   - 这个类也是本人开源的原因

#### 使用说明

1. `go get github.com/fishjam/go-library`


### 注意事项
- 1. 本库中尽量不适用任何第三方的库(比如 log, ut 等), 从而减少导入后的影响.
     但由于 go 中没有统一的日志框架,这个想法可能会比较难满足.