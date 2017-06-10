#include "textflag.h"

TEXT _start(SB),NOSPLIT,$-4
    B   _rt0_arm_vita(SB)

TEXT _rt0_arm_vita(SB),NOSPLIT,$-4
    MOVW	(R13), R0		// argc
    MOVW	$4(R13), R1		// argv
    MOVM.DB.W [R0-R1], (R13)
	B	main(SB)

TEXT main(SB),NOSPLIT,$0
	B	runtime·rt0_go(SB)
