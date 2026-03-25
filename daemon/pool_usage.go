// daemon/pool_usage.go
// ============================================================================
// Pool usage check + Docker engine control
//
// Filosofía:
//   - Storage no toca Docker. Nunca.
//   - Antes de destruir/wipear un pool, la UI pregunta al backend
//     si hay servicios activos usando ese pool.
//   - El usuario para los servicios manualmente desde la app.
//   - Docker es una app independiente con start/stop propio.
// ============================================================================

package main

import (
	"fmt"
	"net/http"
	"strings"
)

// ═══════════════════════════════════════════════════════════════════════════
// checkPoolInUse — qué está usando un pool antes de tocarlo
// ═══════════════════════════════════════════════════════════════════════════

type PoolUsageInfo struct {
	InUse      bool     `json:"inUse"`
	Containers []string `json:"containers"` // nombres de contenedores Docker activos
	Processes  []string `json:"processes"`  // procesos con ficheros abiertos en el pool
	Shares     []string `json:"shares"`     // shares activos en el pool
	Warning    string   `json:"warning"`    // mensaje para el usuario
}

// checkPoolInUse comprueba qué servicios están usando un pool dado su mountPoint.
// No toca nada, solo observa.
func checkPoolInUse(mountPoint string) PoolUsageInfo {
	info := PoolUsageInfo{}

	// ── 1. Contenedores Docker con mounts en este pool ──
	if isDockerInstalledGo() {
		out, ok := run("docker ps --format '{{.Names}}\t{{.Mounts}}' 2>/dev/null")
		if ok && strings.TrimSpace(out) != "" {
			for _, line := range strings.Split(out, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				parts := strings.SplitN(line, "\t", 2)
				name := parts[0]
				mounts := ""
				if len(parts) > 1 {
					mounts = parts[1]
				}
				if strings.Contains(mounts, mountPoint) {
					info.Containers = append(info.Containers, name)
				}
			}
		}
	}

	// ── 2. Procesos con ficheros abiertos en el pool (fuser -v) ──
	fuserOut, _ := run(fmt.Sprintf("fuser -v %s 2>/dev/null | awk '{print $1}' | sort -u", mountPoint))
	if strings.TrimSpace(fuserOut) != "" {
		for _, pid := range strings.Fields(fuserOut) {
			// Obtener nombre del proceso
			comm, _ := run(fmt.Sprintf("cat /proc/%s/comm 2>/dev/null", pid))
			comm = strings.TrimSpace(comm)
			if comm != "" && comm != "nimos-daemon" {
				info.Processes = append(info.Processes, fmt.Sprintf("%s (pid %s)", comm, pid))
			}
		}
	}

	// ── 3. Shares activos en este pool ──
	shares, _ := dbSharesList()
	for _, s := range shares {
		sharePath, _ := s["path"].(string)
		shareName, _ := s["name"].(string)
		if mountPoint != "" && strings.HasPrefix(sharePath, mountPoint) {
			info.Shares = append(info.Shares, shareName)
		}
	}

	// ── Construir warning para el usuario ──
	info.InUse = len(info.Containers) > 0 || len(info.Processes) > 0

	if len(info.Containers) > 0 {
		info.Warning = fmt.Sprintf(
			"Los siguientes contenedores Docker están usando este pool: %s. "+
				"Para los servicios desde la app Docker antes de continuar.",
			strings.Join(info.Containers, ", "),
		)
	} else if len(info.Processes) > 0 {
		info.Warning = fmt.Sprintf(
			"Hay procesos usando este pool: %s. Ciérralos antes de continuar.",
			strings.Join(info.Processes, ", "),
		)
	}

	return info
}

// ═══════════════════════════════════════════════════════════════════════════
// Docker engine control — start/stop/status como app independiente
// ═══════════════════════════════════════════════════════════════════════════

func dockerEngineStatus() map[string]interface{} {
	out, _ := run("systemctl is-active docker 2>/dev/null")
	active := strings.TrimSpace(out) == "active"

	enabled := false
	if out2, _ := run("systemctl is-enabled docker 2>/dev/null"); strings.TrimSpace(out2) == "enabled" {
		enabled = true
	}

	installed := isDockerInstalledGo()

	result := map[string]interface{}{
		"installed": installed,
		"running":   active,
		"enabled":   enabled,
	}

	if active {
		// Info básica de Docker
		version, _ := run("docker version --format '{{.Server.Version}}' 2>/dev/null")
		containers, _ := run("docker ps -q 2>/dev/null | wc -l")
		result["version"] = strings.TrimSpace(version)
		result["runningContainers"] = strings.TrimSpace(containers)
	}

	return result
}

func dockerEngineStart() map[string]interface{} {
	// Verificar que el pool primario está montado
	conf := getStorageConfigFull()
	primaryPool, _ := conf["primaryPool"].(string)
	if primaryPool != "" {
		primaryMount := nimbusPoolsDir + "/" + primaryPool
		if !isMounted(primaryMount) {
			return map[string]interface{}{
				"error": fmt.Sprintf(
					"El pool primario '%s' no está montado. Monta el pool antes de arrancar Docker.",
					primaryPool,
				),
			}
		}
	}

	out, ok := run("systemctl start docker 2>&1")
	if !ok {
		return map[string]interface{}{"error": fmt.Sprintf("No se pudo arrancar Docker: %s", out)}
	}

	logMsg("Docker engine started by user")
	return map[string]interface{}{"ok": true, "message": "Docker arrancado correctamente"}
}

func dockerEngineStop() map[string]interface{} {
	// Informar cuántos contenedores se van a parar
	containersOut, _ := run("docker ps --format '{{.Names}}' 2>/dev/null")
	var running []string
	for _, c := range strings.Split(strings.TrimSpace(containersOut), "\n") {
		if c = strings.TrimSpace(c); c != "" {
			running = append(running, c)
		}
	}

	// Parar Docker (systemd para todos los contenedores automáticamente)
	out, ok := run("systemctl stop docker 2>&1")
	if !ok {
		return map[string]interface{}{"error": fmt.Sprintf("No se pudo parar Docker: %s", out)}
	}

	logMsg("Docker engine stopped by user (%d containers were running)", len(running))
	return map[string]interface{}{
		"ok":      true,
		"message": fmt.Sprintf("Docker parado. %d contenedores detenidos.", len(running)),
		"stoppedContainers": running,
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// HTTP handlers
// ═══════════════════════════════════════════════════════════════════════════

// handlePoolUsageCheck — GET /api/storage/pool/{name}/check-usage
func handlePoolUsageCheck(w http.ResponseWriter, r *http.Request, poolName string) {
	conf := getStorageConfigFull()
	confPools, _ := conf["pools"].([]interface{})

	var mountPoint string
	for _, p := range confPools {
		pm, _ := p.(map[string]interface{})
		if n, _ := pm["name"].(string); n == poolName {
			mountPoint, _ = pm["mountPoint"].(string)
			break
		}
	}

	if mountPoint == "" {
		jsonError(w, fmt.Sprintf("Pool '%s' not found", poolName), http.StatusNotFound)
		return
	}

	info := checkPoolInUse(mountPoint)
	jsonOK(w, info)
}

// handleDockerEngine — /api/docker/engine
func handleDockerEngine(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	if session == nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "GET":
		jsonOK(w, dockerEngineStatus())

	case "POST":
		var body map[string]interface{}
		parseBody(r, &body)
		action, _ := body["action"].(string)

		switch action {
		case "start":
			jsonOK(w, dockerEngineStart())
		case "stop":
			jsonOK(w, dockerEngineStop())
		default:
			jsonError(w, "action must be 'start' or 'stop'", http.StatusBadRequest)
		}

	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
