# go-library

### Some go common functions
  - verify: help functions to handle go error: CSTD(Code Self Test  Development)
    - refer CSTD[doc/使用CSTD技术轻松编写0Bug的代码.doc] , only Chinese
    - example: enable `not_exist` in [virtual_writer_test.go](mime/multipart/virtual_writer_test.go), and can check the error code place and reason
  - flog: simple log wrapper used in verify, user can customize it
  - mime/multipart/VirtualWriter: 
    - similar as go multipart.Writer, but can support upload large files(4G+) with small memory consume 

### Notice
  - 1. Don't use any other 3rd library(example: log,ut) with minimum import impact.

### Usage
  - `go get github.com/fishjam/go-library`