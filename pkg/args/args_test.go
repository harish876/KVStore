package args_test

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/pkg/args"
)

func TestArgs(t *testing.T) {
	args := args.ParseArgs()
	fmt.Println(args)
}
