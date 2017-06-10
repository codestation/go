#include "textflag.h"

// This is only valid for ARMv6+, however, NaCl/ARM is only defined
// for ARMv7A anyway.
TEXT runtime∕internal∕atomic·Cas(SB),NOSPLIT,$0
	B	runtime∕internal∕atomic·armcas(SB)

TEXT runtime∕internal∕atomic·Casp1(SB),NOSPLIT,$0
	B	runtime∕internal∕atomic·Cas(SB)
