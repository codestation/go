package runtime

func netpollinit() {
}

func netpolldescriptor() uintptr {
	return ^uintptr(0)
}

func netpollopen(fd uintptr, pd *pollDesc) int32 {
	return 0
}

func netpollclose(fd uintptr) int32 {
	return 0
}

func netpollarm(pd *pollDesc, mode int) {
}

func netpoll(block bool) *g {
	return nil
}
