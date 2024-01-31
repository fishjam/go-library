# go-library

[![Go Reference](https://pkg.go.dev/badge/github.com/fishjam/go-library.svg)](https://pkg.go.dev/github.com/fishjam/go-library)

### 中文用户请看 [这里](README_zh.md)

### Usage
- `go get github.com/fishjam/go-library`

### Some go common functions
  - verify: help functions to handle go error
    - example: enable `not_exist` in [virtual_writer_test.go](mime/multipart/virtual_writer_test.go), and can check the error code place and reason
  - flog: simple log wrapper used in verify, user need customize it by call `SetLoggerFactory` 
  - mime/multipart/VirtualWriter: 
    - similar as go multipart.Writer, but can support upload large files(4G+) with small memory consume 

### Notice
  - 1. Don't use any other 3rd library(example: log,ut) with minimum import impact.


