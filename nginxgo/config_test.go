package nginxgo

import (
	"testing"
)

func TestConfig(t *testing.T) {
	readConfigFromFile("../configs/config.cfg")
}
