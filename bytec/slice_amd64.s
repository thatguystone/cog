#include "textflag.h"

TEXT ·log2(SB),NOSPLIT,$0-16
	XORQ	AX, AX	// Special case for when == 0
	BSRQ 	n+0(FP), AX
	MOVQ 	AX, ret+8(FP)
	RET
