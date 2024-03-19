package utils

import (
	"fmt"
	"testing"
)

func Test_FieldExistsInFile(t *testing.T) {
	flag, _ := FieldExistsInFile("/Users/zxj/test.sh", []string{"aRETfgfgfgERT"})
	fmt.Printf("%t", flag)
}
