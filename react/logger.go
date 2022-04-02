package react

import (
	"log"
	"os"
)

// Emits logs to this logger. Feel free to change it.
var Logger = log.New(os.Stderr, "react", log.LstdFlags|log.Lshortfile)
