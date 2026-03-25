// daemon/pool_overlays.go
// ============================================================================
// unmountPoolOverlays — desmonta submounts de un pool sin tocar Docker
// ============================================================================

package main

import (
	"fmt"
	"strings"
)

// unmountPoolOverlays desmonta los submounts que cuelgan de un pool
// (overlays de Docker, bind mounts, etc.) sin tocar el proceso Docker
// ni sus configuraciones.
//
// Se llama desde destroyPoolZfs y destroyPoolBtrfs en lugar del bloque
// que antes mataba Docker por la fuerza.
func unmountPoolOverlays(mountPoint string) {
	mountsOut, _ := run(fmt.Sprintf("findmnt -rn -o TARGET %s 2>/dev/null", mountPoint))
	mounts := strings.Split(strings.TrimSpace(mountsOut), "\n")

	// Desmontar en orden inverso — hijos antes que padre
	for i := len(mounts) - 1; i >= 0; i-- {
		m := strings.TrimSpace(mounts[i])
		if m == "" || m == mountPoint {
			continue // no tocar el pool en sí
		}
		logMsg("unmountPoolOverlays: unmounting submount %s", m)
		run(fmt.Sprintf("umount -l %s 2>/dev/null || true", m))
	}
}
