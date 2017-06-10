#include "textflag.h"

TEXT runtime·exit(SB),NOSPLIT,$0
	RET

TEXT runtime·exit1(SB),NOSPLIT,$0
	RET

TEXT runtime·open(SB),NOSPLIT,$0
	RET

TEXT runtime·closefd(SB),NOSPLIT,$0
	RET

TEXT runtime·read(SB),NOSPLIT,$0
	RET

// func naclWrite(fd int, b []byte) int
TEXT syscall·naclWrite(SB),NOSPLIT,$0
	RET

TEXT runtime·write(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_exception_stack(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_exception_handler(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_sem_create(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_sem_wait(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_sem_post(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_mutex_create(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_mutex_lock(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_mutex_trylock(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_mutex_unlock(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_cond_create(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_cond_wait(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_cond_signal(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_cond_broadcast(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_cond_timed_wait_abs(SB),NOSPLIT,$0
	RET

TEXT runtime·nacl_thread_create(SB),NOSPLIT,$0
	RET

TEXT runtime·mstart_nacl(SB),NOSPLIT,$0
	B runtime·mstart(SB)

TEXT runtime·nacl_nanosleep(SB),NOSPLIT,$0
	RET

TEXT runtime·osyield(SB),NOSPLIT,$0
	RET

TEXT runtime·mmap(SB),NOSPLIT,$8
	RET

TEXT runtime·walltime(SB),NOSPLIT,$16
	RET

TEXT syscall·now(SB),NOSPLIT,$0
	B runtime·walltime(SB)

TEXT runtime·nacl_clock_gettime(SB),NOSPLIT,$0
	RET

// int64 nanotime(void) so really
// void nanotime(int64 *nsec)
TEXT runtime·nanotime(SB),NOSPLIT,$16
	RET

TEXT runtime·sigtramp(SB),NOSPLIT,$80
	RET

TEXT runtime·nacl_sysinfo(SB),NOSPLIT,$16
	RET

// func getRandomData([]byte)
TEXT runtime·getRandomData(SB),NOSPLIT,$0-12
	RET

TEXT ·publicationBarrier(SB),NOSPLIT,$-4-0
	B	runtime·armPublicationBarrier(SB)
