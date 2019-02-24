package wheel

import "time"

func CheapNow() time.Time {
	return cheapTime()
}

func CheapNowUnix() int64 {
	return cheapTime().Unix()
}

func CheapNowNano() int64 {
	return cheapTime().UnixNano()
}

func cheapTime() time.Time {
	return defaultWheel().now.Load().(time.Time)
}
