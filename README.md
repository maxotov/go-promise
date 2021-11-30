# go-promise

promise library for Go

### Installation

`go get github.com/maxotov/go-promise`

### Usage

```go
import (
    "context"
	"fmt"
	"github.com/maxotov/go-promise"
	"time"
)

func main() {
    
	promises := promise.NewInt64MultiPromises()
    
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
        defer cancel()
        value, err := promises.WaitForValue(ctx, "111")
        if err != nil {
            panic("timeout during get value number from promise")
        }
        fmt.Printf("value from promise: %d", value)
    }()
    
    promises.Resolve("111", 222)
    time.Sleep(time.Second)
	
}
```
