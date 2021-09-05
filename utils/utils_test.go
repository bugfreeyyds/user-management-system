package utils

import (
    "fmt"
    "testing"
    //"gotest.tools/assert"
)

func Test_Md5String(t *testing.T) {
    s := "abcdefg"
    result := Md5String(s)
    if result != "7ac66c0f148de9519b8bd264312c4d64" {
        t.Error("test failed: ", result)
    }
    s = "123456dd987c"
    result = Md5String(s)
    fmt.Println(result)
    t.Error(result)
//    assert.Equal(t, result, "32232232")
}

func Test_GenerateToken(t *testing.T) {
    s := "abcdefg"
    result := GenerateToken(s)
    if result != "6d020719bc3fcafe85b125501b9f9737" {
        t.Error("test failed: ", result)
    }
}
