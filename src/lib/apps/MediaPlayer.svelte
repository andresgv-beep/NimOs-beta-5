<script>
  import { onMount, onDestroy } from 'svelte';
  import { getToken } from '$lib/stores/auth.js';

  const hdrs = () => ({ 'Authorization': `Bearer ${getToken()}` });
  const token = getToken();

  let shares = [];
  let currentShare = '';
  let currentPath = '/';
  let files = [];
  let loading = false;

  let playerEl;
  let isVideo = false;
  let playing = false;
  let currentFile = null;
  let currentSrc = '';
  let duration = 0;
  let currentTime = 0;
  let volume = 0.8;
  let muted = false;

  let playlist = [];
  let playlistIdx = -1;

  let controlsVisible = true;
  let hideTimer = null;

  const AUDIO_EXT = ['mp3','wav','flac','aac','m4a','ogg','opus','wma'];
  const VIDEO_EXT = ['mp4','webm','mkv','avi','mov','ogv'];
  const MEDIA_EXT = [...AUDIO_EXT, ...VIDEO_EXT];

  function getExt(name) { const d = name.lastIndexOf('.'); return d >= 0 ? name.slice(d+1).toLowerCase() : ''; }
  function isMedia(name)     { return MEDIA_EXT.includes(getExt(name)); }
  function isVideoFile(name) { return VIDEO_EXT.includes(getExt(name)); }
  function streamUrl(share, path) {
    return `/api/files/download?share=${encodeURIComponent(share)}&path=${encodeURIComponent(path)}&token=${encodeURIComponent(token)}`;
  }

  async function loadShares() {
    try {
      const res = await fetch('/api/files', { headers: hdrs() });
      const data = await res.json();
      shares = data.shares || [];
      if (shares.length > 0 && !currentShare) { currentShare = shares[0].name; loadFiles(); }
    } catch {}
  }

  async function loadFiles() {
    if (!currentShare) return;
    loading = true;
    try {
      const res = await fetch(`/api/files?share=${encodeURIComponent(currentShare)}&path=${encodeURIComponent(currentPath)}`, { headers: hdrs() });
      const data = await res.json();
      files = data.files || [];
    } catch {}
    loading = false;
  }

  function enterFolder(name) {
    currentPath = currentPath === '/' ? '/' + name : currentPath + '/' + name;
    loadFiles();
  }

  function goUp() {
    if (currentPath === '/') return;
    const parts = currentPath.split('/').filter(Boolean); parts.pop();
    currentPath = parts.length ? '/' + parts.join('/') : '/';
    loadFiles();
  }

  function selectShare(name) { currentShare = name; currentPath = '/'; loadFiles(); }

  function playFile(file) {
    const path = currentPath === '/' ? '/' + file.name : currentPath + '/' + file.name;
    currentFile = file;
    currentSrc = streamUrl(currentShare, path);
    isVideo = isVideoFile(file.name);
    playing = true;
    playlist = files.filter(f => isMedia(f.name));
    playlistIdx = playlist.findIndex(f => f.name === file.name);
    if (playerEl) { playerEl.src = currentSrc; playerEl.load(); playerEl.play().catch(() => {}); }
    scheduleHide();
  }

  function playNext() {
    if (!playlist.length) return;
    playlistIdx = (playlistIdx + 1) % playlist.length;
    playFile(playlist[playlistIdx]);
  }

  function playPrev() {
    if (!playlist.length) return;
    if (currentTime > 3) { if (playerEl) playerEl.currentTime = 0; return; }
    playlistIdx = (playlistIdx - 1 + playlist.length) % playlist.length;
    playFile(playlist[playlistIdx]);
  }

  function togglePlay() {
    if (!playerEl) return;
    if (playerEl.paused) { playerEl.play().catch(() => {}); scheduleHide(); }
    else { playerEl.pause(); showControls(); }
  }

  function seek(e) {
    if (!playerEl || !duration) return;
    const rect = e.currentTarget.getBoundingClientRect();
    playerEl.currentTime = ((e.clientX - rect.left) / rect.width) * duration;
  }

  function setVol(e) {
    const rect = e.currentTarget.getBoundingClientRect();
    volume = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
    if (playerEl) playerEl.volume = volume;
    muted = volume === 0;
  }

  function toggleMute() { muted = !muted; if (playerEl) playerEl.muted = muted; }

  function toggleFullscreen() {
    const el = document.querySelector('.mp-inner');
    if (!el) return;
    if (!document.fullscreenElement) el.requestFullscreen().catch(() => {});
    else document.exitFullscreen();
  }

  function showControls() { clearTimeout(hideTimer); controlsVisible = true; }

  function scheduleHide() {
    clearTimeout(hideTimer);
    controlsVisible = true;
    hideTimer = setTimeout(() => { if (playing) controlsVisible = false; }, 3000);
  }

  function onMouseMove() { if (playing) scheduleHide(); else showControls(); }

  function fmtTime(s) {
    if (!s || isNaN(s)) return '0:00';
    return `${Math.floor(s/60)}:${Math.floor(s%60).toString().padStart(2,'0')}`;
  }

  function fmtSize(b) {
    if (!b) return '';
    if (b >= 1e9) return (b/1e9).toFixed(1)+' GB';
    if (b >= 1e6) return (b/1e6).toFixed(1)+' MB';
    return (b/1e3).toFixed(0)+' KB';
  }

  $: breadcrumbs = currentPath === '/' ? [] : currentPath.split('/').filter(Boolean);

  onMount(loadShares);
  onDestroy(() => clearTimeout(hideTimer));
</script>

<div class="mp-root">

  <!-- ── SIDEBAR ── -->
  <div class="mp-sidebar">
    <div class="mp-sidebar-header">
      <svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>
      <span class="mp-title">Media</span>
    </div>

    <div class="mp-section-header">
      <span class="mp-section-label">Cola</span>
      <button class="mp-add-btn" title="Añadir a la lista">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
      </button>
    </div>

    <div class="mp-playlist">
      {#each playlist as item, i}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div class="mp-pl-item" class:active={i === playlistIdx} on:click={() => { playlistIdx = i; playFile(item); }}>
          {#if isVideoFile(item.name)}
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><rect x="2" y="2" width="20" height="20" rx="2"/><polygon points="10 8 16 12 10 16 10 8"/></svg>
          {:else}
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/></svg>
          {/if}
          <span class="mp-pl-name">{item.name}</span>
          {#if i === playlistIdx}
            <div class="mp-pl-playing"><span></span><span></span><span></span></div>
          {/if}
        </div>
      {/each}
      {#if playlist.length === 0}
        <div class="mp-pl-empty">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/></svg>
          <span>Sin archivos en cola</span>
        </div>
      {/if}
    </div>

  </div>

  <!-- ── INNER WRAP (patrón NimOS) ── -->
  <div class="mp-inner-wrap">
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="mp-inner" on:mousemove={onMouseMove}>

      <!-- Titlebar con breadcrumb -->
      <div class="mp-titlebar" class:hidden={currentFile && isVideo && currentSrc}>
        <span class="mp-tb-title">Media</span>
        <span class="mp-tb-sep">—</span>
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <span class="mp-bc-root" on:click={() => { currentPath='/'; loadFiles(); }}>{currentShare || '—'}</span>
        {#each breadcrumbs as crumb, i}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" style="width:9px;height:9px;color:var(--text-3);flex-shrink:0"><polyline points="9 18 15 12 9 6"/></svg>
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <span class="mp-bc-crumb" on:click={() => { currentPath='/'+breadcrumbs.slice(0,i+1).join('/'); loadFiles(); }}>{crumb}</span>
        {/each}
      </div>

      <!-- Contenido: vídeo, audio o pantalla vacía -->
      <div class="mp-content">
        {#if currentFile && isVideo && currentSrc}
          <div class="mp-video-wrap">
            <!-- svelte-ignore a11y_media_has_caption -->
            <video
              bind:this={playerEl}
              src={currentSrc}
              bind:duration bind:currentTime bind:volume bind:muted
              on:ended={playNext}
              on:play={() => playing = true}
              on:pause={() => playing = false}
              class="mp-video"
            ></video>
          </div>
        {:else if currentFile && !isVideo && currentSrc}
          <!-- Audio: pantalla de now playing -->
          <div class="mp-audio-screen">
            <div class="mp-audio-art">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round"><path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/></svg>
            </div>
            <div class="mp-audio-name">{currentFile.name}</div>
            <div class="mp-audio-path">{currentShare}</div>
          </div>
          <!-- svelte-ignore a11y_media_has_caption -->
          <audio bind:this={playerEl} src={currentSrc}
            bind:duration bind:currentTime bind:volume bind:muted
            on:ended={playNext}
            on:play={() => playing = true}
            on:pause={() => playing = false}></audio>
        {:else}
          <!-- Empty state -->
          <div class="mp-empty-screen">
            <div class="mp-empty-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round"><polygon points="5 3 19 12 5 21 5 3"/></svg>
            </div>
            <div class="mp-empty-title">Sin reproducción</div>
            <div class="mp-empty-desc">Usa el botón + para añadir archivos a la cola</div>
          </div>
        {/if}

        <!-- Controles flotantes -->
        {#if currentFile}
          <div class="mp-controls" class:hidden={!controlsVisible}>
            <div class="mp-progress-row">
              <span class="mp-time">{fmtTime(currentTime)}</span>
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <div class="mp-progress" on:click={seek}>
                <div class="mp-progress-fill" style="width:{duration ? (currentTime/duration)*100 : 0}%">
                  <div class="mp-progress-thumb"></div>
                </div>
              </div>
              <span class="mp-time">{fmtTime(duration)}</span>
            </div>
            <div class="mp-btns-row">
              <div class="mp-now">
                <div class="mp-now-art" class:video={isVideo}>
                  {#if isVideo}
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><rect x="2" y="2" width="20" height="20" rx="2"/><polygon points="10 8 16 12 10 16 10 8"/></svg>
                  {:else}
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/></svg>
                  {/if}
                </div>
                <div class="mp-now-info">
                  <div class="mp-now-name">{currentFile.name}</div>
                  <div class="mp-now-path">{currentShare}{currentPath === '/' ? '' : currentPath}</div>
                </div>
              </div>
              <div class="mp-transport">
                <button class="mp-btn" on:click={playPrev}>
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="19 20 9 12 19 4 19 20"/><line x1="5" y1="19" x2="5" y2="5"/></svg>
                </button>
                <button class="mp-btn play" on:click={togglePlay}>
                  {#if playing}
                    <svg viewBox="0 0 24 24" fill="currentColor"><rect x="6" y="4" width="4" height="16" rx="1"/><rect x="14" y="4" width="4" height="16" rx="1"/></svg>
                  {:else}
                    <svg viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>
                  {/if}
                </button>
                <button class="mp-btn" on:click={playNext}>
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 4 15 12 5 20 5 4"/><line x1="19" y1="5" x2="19" y2="19"/></svg>
                </button>
              </div>
              <div class="mp-right">
                <div class="mp-vol-wrap">
                  <button class="mp-btn small" on:click={toggleMute}>
                    {#if muted || volume === 0}
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/><line x1="23" y1="9" x2="17" y2="15"/><line x1="17" y1="9" x2="23" y2="15"/></svg>
                    {:else if volume < 0.5}
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/><path d="M15.54 8.46a5 5 0 0 1 0 7.07"/></svg>
                    {:else}
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/><path d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"/></svg>
                    {/if}
                  </button>
                  <!-- svelte-ignore a11y_click_events_have_key_events -->
                  <!-- svelte-ignore a11y_no_static_element_interactions -->
                  <div class="mp-vol-track" on:click={setVol}>
                    <div class="mp-vol-fill" style="width:{muted ? 0 : volume*100}%"></div>
                  </div>
                </div>
                <button class="mp-btn fullscreen" on:click={toggleFullscreen} title="Pantalla completa">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/></svg>
                </button>
              </div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Statusbar -->
      <div class="mp-statusbar" class:hidden={currentFile && isVideo && currentSrc}>
        <div class="mp-status-dot"></div>
        <span>NimOS Beta 5</span>
      </div>
    </div>
  </div>
</div>

<style>
  .mp-root { width:100%; height:100%; display:flex; overflow:hidden; font-family:'Inter',-apple-system,sans-serif; color:var(--text-1); }

  /* ── Sidebar ── */
  .mp-sidebar { width:200px; flex-shrink:0; padding:12px 8px; background:var(--bg-sidebar); display:flex; flex-direction:column; overflow:hidden; }
  .mp-sidebar-header { display:flex; align-items:center; gap:8px; padding:28px 10px 14px; color:var(--text-1); }
  .mp-title { font-size:15px; font-weight:600; }
  .mp-section-header { display:flex; align-items:center; justify-content:space-between; padding:4px 8px; }
  .mp-section-label { font-size:9px; font-weight:600; color:var(--text-3); text-transform:uppercase; letter-spacing:.08em; }
  .mp-add-btn { width:18px; height:18px; border-radius:4px; border:1px solid var(--border); background:transparent; color:var(--text-3); cursor:pointer; display:flex; align-items:center; justify-content:center; transition:all .15s; }
  .mp-add-btn svg { width:10px; height:10px; }
  .mp-add-btn:hover { color:var(--text-1); border-color:var(--border-hi); background:var(--ibtn-bg); }
  .mp-playlist { display:flex; flex-direction:column; gap:1px; overflow-y:auto; flex:1; margin-top:2px; }
  .mp-playlist::-webkit-scrollbar { width:3px; }
  .mp-playlist::-webkit-scrollbar-thumb { background:rgba(128,128,128,0.15); border-radius:2px; }
  .mp-pl-item { display:flex; align-items:center; gap:7px; padding:6px 8px; border-radius:6px; font-size:11px; color:var(--text-3); cursor:pointer; transition:all .1s; }
  .mp-pl-item svg { width:11px; height:11px; flex-shrink:0; opacity:.6; }
  .mp-pl-item:hover { background:rgba(128,128,128,0.06); color:var(--text-2); }
  .mp-pl-item.active { color:var(--accent); background:var(--active-bg); }
  .mp-pl-item.active svg { opacity:1; }
  .mp-pl-name { flex:1; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
  .mp-pl-playing { display:flex; align-items:flex-end; gap:2px; height:12px; flex-shrink:0; }
  .mp-pl-playing span { width:2px; border-radius:1px; background:var(--accent); animation:bar-bounce .8s ease-in-out infinite; }
  .mp-pl-playing span:nth-child(1) { height:5px; animation-delay:0s; }
  .mp-pl-playing span:nth-child(2) { height:9px; animation-delay:.2s; }
  .mp-pl-playing span:nth-child(3) { height:6px; animation-delay:.4s; }
  @keyframes bar-bounce { 0%,100%{transform:scaleY(0.4)} 50%{transform:scaleY(1)} }
  .mp-pl-empty { display:flex; flex-direction:column; align-items:center; gap:8px; padding:24px 8px; opacity:.4; }
  .mp-pl-empty svg { width:22px; height:22px; }
  .mp-pl-empty span { font-size:10px; color:var(--text-3); text-align:center; }


  /* ── Inner wrap — patrón NimOS ── */
  .mp-inner-wrap { flex:1; padding:8px; display:flex; }
  .mp-inner {
    flex:1; border-radius:10px; border:1px solid var(--border);
    background:var(--bg-inner); display:flex; flex-direction:column;
    overflow:hidden; position:relative;
  }

  /* Titlebar */
  .mp-titlebar.hidden { display:none; }
  .mp-statusbar.hidden { display:none; }
  .mp-titlebar {
    display:flex; align-items:center; gap:6px;
    padding:14px 18px 13px;
    background:var(--bg-bar); flex-shrink:0;
    border-bottom:1px solid var(--border);
  }
  .mp-tb-title { font-size:13px; font-weight:600; color:var(--text-1); }
  .mp-tb-sep { font-size:11px; color:var(--text-3); }
  .mp-bc-root { font-size:11px; color:var(--text-3); cursor:pointer; transition:color .1s; font-family:'DM Mono',monospace; }
  .mp-bc-root:hover { color:var(--text-2); }
  .mp-bc-crumb { font-size:10px; color:var(--text-3); cursor:pointer; font-family:'DM Mono',monospace; transition:color .1s; }
  .mp-bc-crumb:hover { color:var(--text-2); }

  /* Content area */
  .mp-content { flex:1; position:relative; overflow:hidden; }

  /* Video */
  .mp-video-wrap { position:absolute; inset:0; display:flex; align-items:center; justify-content:center; background:#000; }
  .mp-video { width:100%; height:100%; object-fit:contain; }

  /* Pantallas de contenido */
  .mp-empty-screen { position:absolute; inset:0; display:flex; flex-direction:column; align-items:center; justify-content:center; gap:12px; }
  .mp-empty-icon { width:64px; height:64px; border-radius:16px; background:rgba(124,111,255,0.08); border:1px solid rgba(124,111,255,0.12); display:flex; align-items:center; justify-content:center; }
  .mp-empty-icon svg { width:28px; height:28px; color:var(--accent); opacity:.4; }
  .mp-empty-title { font-size:13px; font-weight:600; color:var(--text-2); }
  .mp-empty-desc { font-size:11px; color:var(--text-3); }

  .mp-audio-screen { position:absolute; inset:0; display:flex; flex-direction:column; align-items:center; justify-content:center; gap:14px; padding-bottom:80px; }
  .mp-audio-art { width:120px; height:120px; border-radius:20px; background:rgba(124,111,255,0.12); border:1px solid rgba(124,111,255,0.20); display:flex; align-items:center; justify-content:center; }
  .mp-audio-art svg { width:52px; height:52px; color:var(--accent); opacity:.6; }
  .mp-audio-name { font-size:14px; font-weight:600; color:var(--text-1); max-width:400px; text-align:center; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
  .mp-audio-path { font-size:10px; color:var(--text-3); font-family:'DM Mono',monospace; }

  .mp-loading { display:flex; justify-content:center; padding:30px; }
  .spinner { width:20px; height:20px; border-radius:50%; border:2px solid rgba(255,255,255,0.08); border-top-color:var(--accent); animation:spin .7s linear infinite; }
  @keyframes spin { to{transform:rotate(360deg)} }

  /* Controles flotantes */
  .mp-controls { position:absolute; bottom:0; left:0; right:0; padding:20px 20px 16px; background:linear-gradient(to top,rgba(0,0,0,0.88) 0%,rgba(0,0,0,0.45) 60%,transparent 100%); display:flex; flex-direction:column; gap:10px; transition:opacity .35s ease; z-index:10; border-radius:0 0 10px 10px; }
  .mp-controls.hidden { opacity:0; pointer-events:none; }
  .mp-progress-row { display:flex; align-items:center; gap:8px; }
  .mp-time { font-size:10px; color:rgba(255,255,255,0.55); font-family:'DM Mono',monospace; min-width:32px; text-align:center; }
  .mp-progress { flex:1; height:3px; border-radius:2px; background:rgba(255,255,255,0.15); cursor:pointer; position:relative; transition:height .15s; }
  .mp-progress:hover { height:5px; }
  .mp-progress-fill { height:100%; border-radius:2px; background:linear-gradient(90deg,var(--accent),var(--accent2)); position:relative; }
  .mp-progress-thumb { position:absolute; right:-5px; top:50%; transform:translateY(-50%) scale(0); width:11px; height:11px; border-radius:50%; background:#fff; transition:transform .15s; }
  .mp-progress:hover .mp-progress-thumb { transform:translateY(-50%) scale(1); }
  .mp-btns-row { display:flex; align-items:center; gap:4px; }
  .mp-now { display:flex; align-items:center; gap:9px; flex:1; min-width:0; }
  .mp-now-art { width:34px; height:34px; border-radius:7px; flex-shrink:0; background:rgba(124,111,255,0.20); border:1px solid rgba(124,111,255,0.30); display:flex; align-items:center; justify-content:center; }
  .mp-now-art svg { width:15px; height:15px; color:var(--accent); }
  .mp-now-art.video { background:rgba(96,165,250,0.15); border-color:rgba(96,165,250,0.25); }
  .mp-now-art.video svg { color:var(--blue); }
  .mp-now-info { overflow:hidden; min-width:0; }
  .mp-now-name { font-size:12px; font-weight:600; color:#fff; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
  .mp-now-path { font-size:9px; color:rgba(255,255,255,0.4); font-family:'DM Mono',monospace; }
  .mp-transport { display:flex; align-items:center; gap:6px; flex:1; justify-content:center; margin-right:80px; }
  .mp-transport .mp-btn svg { width:17px; height:17px; }
  .mp-btn { width:34px; height:34px; border:none; background:none; color:rgba(255,255,255,0.7); cursor:pointer; border-radius:8px; display:flex; align-items:center; justify-content:center; transition:all .12s; }
  .mp-btn svg { width:15px; height:15px; }
  .mp-btn:hover { background:rgba(255,255,255,0.10); color:#fff; }
  .mp-btn.play { width:42px; height:42px; border-radius:50%; background:rgba(255,255,255,0.15); backdrop-filter:blur(8px); border:1px solid rgba(255,255,255,0.25); color:#fff; }
  .mp-btn.play svg { width:18px; height:18px; }
  .mp-btn.play:hover { background:rgba(255,255,255,0.25); }
  .mp-btn.small { width:28px; height:28px; }
  .mp-btn.small svg { width:13px; height:13px; }
  .mp-right { display:flex; align-items:center; gap:4px; flex-shrink:0; }
  .mp-vol-wrap { display:flex; align-items:center; gap:6px; }
  .mp-vol-track { width:68px; height:3px; border-radius:2px; background:rgba(255,255,255,0.15); cursor:pointer; transition:height .15s; }
  .mp-vol-track:hover { height:5px; }
  .mp-vol-fill { height:100%; border-radius:2px; background:rgba(255,255,255,0.7); }

  /* Statusbar */
  .mp-statusbar { display:flex; align-items:center; gap:10px; padding:8px 16px; border-top:1px solid var(--border); background:var(--bg-bar); flex-shrink:0; font-size:10px; color:var(--text-3); border-radius:0 0 10px 10px; font-family:'DM Mono',monospace; }
  .mp-status-dot { width:6px; height:6px; border-radius:50%; background:var(--green); box-shadow:0 0 4px rgba(74,222,128,0.6); }
</style>
