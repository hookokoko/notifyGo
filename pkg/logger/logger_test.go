package logger

import "testing"

func Test_Base(t *testing.T) {
	Default().Info("hello world", String("h", "231"))
}
