package stringsutil_test

import (
    "fmt"

    "github.com/mouuii/stringsutil"
)

func ExampleReverse() {
    fmt.Println(stringsutil.Reverse("hello"))
    // Output: olleh
}