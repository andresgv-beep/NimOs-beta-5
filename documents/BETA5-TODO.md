# NimOS Beta 5 — Estado y Pendientes

## Fecha: 2026-03-15
## Repo: NimOs-beta-5 (SvelteKit)

---

## ✅ COMPLETADO

### Core OS (Svelte)
- [x] Stores: auth.js, theme.js, windows.js
- [x] Login + SetupWizard
- [x] Desktop + Taskbar + WindowFrame
- [x] Controles de ventana propios (3 líneas colores)
- [x] Drag + resize + minimize + maximize ventanas
- [x] App loader dinámico en WindowFrame (files, settings, torrent, appstore)

### Apps portadas (UI)
- [x] FileManager (diseño mockup HTML)
- [x] Settings (por Sonnet)
- [x] NimTorrent (por Sonnet)
- [x] AppStore — carga catálogo desde GitHub, muestra apps con screenshots

### Backend (daemon Go)
- [x] Docker install con timeout 300s + logging
- [x] ProtectSystem=false en service (para poder instalar Docker)
- [x] storage.go — appendFstab con fallback device path si UUID vacío
- [x] AppStore endpoints corregidos: /api/docker/stack, /api/docker/installed-apps, DELETE /api/docker/stack/{id}
- [x] AppStore.svelte — response parsing fijo (data.apps vs array directo)

---

## ❌ PENDIENTE — FLUJO DOCKER/APPS (PRIORIDAD ALTA)

### 1. Launcher (cajón de apps)
**En beta 4:** `Launcher.jsx` — carga `/api/docker/installed-apps`, muestra Docker apps junto a las del sistema en un grid. Al click abre ventana con iframe.

**En beta 5:** NO EXISTE. Solo hay Taskbar con pinned apps.

**Necesita:**
- [ ] Crear `Launcher.svelte` — grid de apps (sistema + Docker instaladas)
- [ ] Botón en Taskbar para abrir Launcher (o click derecho en desktop)
- [ ] Launcher carga `/api/docker/installed-apps` y muestra iconos
- [ ] Click en Docker app → `openWindow(id, size, {isWebApp:true, port, appName})`
- [ ] Click en app del sistema → `openWindow(id)`

### 2. WebApp iframe
**En beta 4:** `WebApp.jsx` — iframe a `http://hostname:port` con loading/error states.

**En beta 5:** NO EXISTE.

**Necesita:**
- [ ] Crear `WebApp.svelte` — iframe a `http://{hostname}:{port}`
- [ ] Pre-check con fetch no-cors antes de mostrar iframe
- [ ] Estados: loading spinner, error con retry + abrir externo, ready con iframe
- [ ] Sandbox: allow-same-origin allow-scripts allow-forms allow-popups

### 3. WindowFrame — soporte WebApp
**En beta 4:** WindowFrame detecta `win.isWebApp` y renderiza `<WebApp>` en vez de app nativa.

**En beta 5:** WindowFrame NO maneja isWebApp.

**Necesita:**
- [ ] Añadir `{:else if win.isWebApp}` en WindowFrame → renderizar WebApp.svelte
- [ ] Pasar `win.webAppPort`, `win.webAppName` al componente

### 4. AppStore install flow completo
**En beta 4:** Install wizard con progreso, procesamiento de env ({RANDOM}, ${CONFIG_PATH}), credenciales post-install.

**En beta 5:** Install básico sin procesamiento de env ni credenciales.

**Necesita:**
- [ ] Procesar env: `{RANDOM}` → password generado, `${CONFIG_PATH}` → path del stack
- [ ] Mostrar progreso durante instalación (descargando, iniciando...)
- [ ] Mostrar credenciales post-install si la app las define en el catálogo
- [ ] Refrescar lista de installed apps después de instalar

### 5. Docker share automático
**En beta 4:** Al instalar Docker se creaba acceso a la carpeta docker del pool.

**En beta 5:** NO se crea share automático.

**Necesita:**
- [ ] En `dockerInstall()` de docker.go: crear share "docker" apuntando a `{pool}/docker`
- [ ] Permisos rw para admin en la carpeta docker y subcarpetas

### 6. Taskbar — Docker apps
**En beta 4:** Taskbar también cargaba `/api/docker/installed-apps` y mostraba iconos de apps corriendo.

**En beta 5:** Taskbar solo muestra pinned apps del sistema.

**Necesita:**
- [ ] Taskbar carga Docker apps instaladas
- [ ] Muestra iconos de apps Docker corriendo al lado de las del sistema

---

## ❌ PENDIENTE — CONTAINERS APP

### 7. Containers manager
**En beta 4:** App Containers mostraba contenedores corriendo con controles (start/stop/restart/logs).

**En beta 5:** NO EXISTE (placeholder "Coming soon").

**Necesita:**
- [ ] Crear `Containers.svelte` 
- [ ] Lista de contenedores desde `/api/docker/status` (containers array)
- [ ] Acciones: start, stop, restart, remove por contenedor
- [ ] Logs: `/api/docker/container/{id}/logs`
- [ ] Estado con colores (running=verde, stopped=rojo)

---

## ❌ PENDIENTE — POLISH Y OTROS

### 8. Temas y scaling
- [ ] CSS tokens globales para los 3 temas (midnight/dark/light) en app.css
- [ ] Accent color dinámico desde theme store
- [ ] Scaling responsive para 1080p / 1440p / 4K

### 9. Widgets
- [ ] Clock, DiskPool, SystemMonitor, Network, NimTorrent widgets
- [ ] Widget grid en desktop

### 10. Media Player
- [ ] Portar MediaPlayer a Svelte

### 11. System Monitor
- [ ] Portar SystemMonitor a Svelte

### 12. Context menus
- [ ] Click derecho en desktop → menú contextual
- [ ] Click derecho en archivos → opciones (copiar, mover, renombrar, borrar)

### 13. Limpieza repo
- [ ] Borrar `src/components/` (duplicado viejo)
- [ ] Actualizar URLs en catalog.json (nimbusos-appstore → NimOs-appstore)
- [ ] Borrar archivos sobrantes de la raíz (docker.go, http.go viejos)

---

## ARCHIVOS CLAVE — Referencia

### Beta 5 (Svelte)
```
src/routes/+page.svelte          — Entry point
src/lib/stores/auth.js           — Auth store
src/lib/stores/theme.js          — Theme/prefs store  
src/lib/stores/windows.js        — Window manager store
src/lib/apps.js                  — App metadata (icons, sizes)
src/lib/components/Desktop.svelte
src/lib/components/Taskbar.svelte
src/lib/components/WindowFrame.svelte
src/lib/components/Login.svelte
src/lib/components/SetupWizard.svelte
src/lib/apps/FileManager.svelte
src/lib/apps/Settings.svelte
src/lib/apps/NimTorrent.svelte
src/lib/apps/AppStore.svelte
```

### Daemon Go
```
daemon/main.go      — Startup, socket, monitoring
daemon/http.go      — HTTP routes
daemon/auth.go      — Login, users, prefs
daemon/docker.go    — Docker install, stacks, containers
daemon/storage.go   — RAID pools, disks, fstab
daemon/files.go     — File manager API
daemon/network.go   — DDNS, SSL, services
daemon/apps.go      — Native/installed apps
daemon/shares.go    — Shared folders
daemon/hardware.go  — CPU/GPU/temps/disks
```

### Endpoints Docker
```
GET  /api/docker/status          — Docker + containers status
POST /api/docker/install         — Install Docker engine
GET  /api/docker/installed-apps  — List installed apps (returns {apps:[...]})
POST /api/docker/stack           — Deploy stack (compose)
DELETE /api/docker/stack/{id}    — Remove stack
GET  /api/docker/containers      — List containers
POST /api/docker/container/{id}/{action} — start/stop/restart
```

### Repo AppStore
```
https://github.com/andresgv-beep/NimOs-appstore
├── catalog.json    — App definitions
├── icons/          — SVG icons
└── screenshots/    — App screenshots
```
