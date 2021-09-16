package consumers

type Consumer func(t string, c []string, rows chan []string)
