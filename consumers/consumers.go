package consumers

import "sync"

type Consumer func(*sync.WaitGroup) func(t string, c []string, rows chan []string)
