# tarutil

压缩和解压缩以 `.tar` 或 `.tar.gz` 为后缀的压缩文件工具包。

## example

```go
import "github.com/treeforest/tarutil"
// ...

// 压缩
tarutil.Archive(srcFileOrDirectory, tarFilePath)

// 解压
tarutil.Extract(tarFilePath, dstDirectory)
```