package debugutil

import (
	"io"
)

/***********************************************************************************************************************
* 注意:
*   1.如果是通过 defer 调用的话, 定位出来的位置通常在调用函数的结束位置, 需要搜索查询确认具体位置(TODO: 或增加 message ?)
*   2.常见错误:
*     fs.ErrClosed <== *fs.PathError(close |0: file already closed)
***********************************************************************************************************************/
func SafeClose(closer io.Closer) {
	if closer != nil {
		_ = VerifyWithConfig(closer.Close(), &Config{
			MoreSkip: 1,
		})
	}
}
