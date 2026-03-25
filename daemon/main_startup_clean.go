// ============================================================================
// PATCH main.go — bloque de startup limpio (sin Docker)
// Sustituir desde "// FIRST: Mount all pools..." hasta startZfsScheduler()
// ============================================================================

	// 1. Montar todos los pools registrados (ZFS, BTRFS, mdraid/ext4)
	//    Esto es TODO lo que hace el daemon en startup respecto a storage.
	//    Docker NO se toca aquí — es responsabilidad del usuario arrancarlo
	//    desde la app una vez los pools estén montados.
	ensurePoolsMounted()

	// 2. Monitoring en background (alertas de espacio, scrub, etc.)
	//    DEBE ir después de ensurePoolsMounted para no borrar
	//    mountpoints válidos en cleanOrphanMountPoints.
	startStorageMonitoring()

	// 3. Scheduler de snapshots ZFS automáticos
	startZfsScheduler()

	// 4. Notificar a systemd que estamos listos.
	//    Con Type=notify en el .service, systemd puede arrancar otros
	//    servicios que dependan de nosotros (nimbusos.service, etc.)
	run("systemd-notify --ready 2>/dev/null || true")

// ============================================================================
// TAMBIÉN ELIMINAR de main.go las llamadas a funciones que ya no existen:
//   - zfsAutoImportOnStartup()   → reemplazada por ensurePoolsMounted()
//   - btrfsAutoMountOnStartup()  → reemplazada por ensurePoolsMounted()
//   - startupStorageAndDocker()  → eliminada, no necesaria
// ============================================================================
