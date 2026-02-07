package config

import "os"

func IsDebug() bool {
	return os.Getenv("TUSK_DEBUG") == "1"
}
