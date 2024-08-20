package debugutil

import (
	"io"
)

/***********************************************************************************************************************
* 注意:
*   1.如果是通过 defer 调用的话, 定位出来的位置通常在调用函数的结束位置, 推荐使用 SafeCloseMsg() 方便定位
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

func SafeCloseMsg(closer io.Closer, msg string) {
	if closer != nil {
		_ = VerifyWithConfig(closer.Close(), &Config{
			MoreSkip: 1,
			Message: msg,
		})
	}
}
