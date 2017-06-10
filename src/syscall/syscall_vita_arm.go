package syscall

type Timespec struct {
	Sec  int64
	Nsec int32
}

type Timeval struct {
	Sec  int64
	Usec int32
}

func setTimespec(sec, nsec int64) Timespec {
	return Timespec{Sec: sec, Nsec: int32(nsec)}
}

func setTimeval(sec, usec int64) Timeval {
	return Timeval{Sec: sec, Usec: int32(usec)}
}
