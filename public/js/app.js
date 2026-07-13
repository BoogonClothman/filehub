// ── state ──────────────────────────────────────────────────────────
const state = {
  path: '',
  entries: [],
  contextTarget: null,   // { name, isDir, preview, path } — set on right-click
  pendingTarget: null,   // survives context menu close until dialog resolves
  currentDialog: null,
};

// ── DOM refs ────────────────────────────────────────────────────────
const $ = (sel, ctx = document) => ctx.querySelector(sel);
const $$ = (sel, ctx = document) => [...ctx.querySelectorAll(sel)];

const el = {
  breadcrumb: $('#breadcrumb'),
  fileList: $('#file-list'),
  emptyHint: $('#empty-hint'),
  loading: $('#loading'),
  errorMsg: $('#error-msg'),
  dropZone: $('#drop-zone'),
  dropHint: $('#drop-hint'),
  contextMenu: $('#context-menu'),
  fileInput: $('#file-input'),

  dialogNewDir: $('#dialog-newdir'),
  inputNewDir: $('#input-newdir'),

  dialogRename: $('#dialog-rename'),
  inputRename: $('#input-rename'),

  dialogDelete: $('#dialog-delete'),
  deleteName: $('#delete-name'),

  dialogPreview: $('#dialog-preview'),
  previewImg: $('#preview-img'),
  previewName: $('#preview-name'),
};

// ── API helpers ─────────────────────────────────────────────────────
async function api(url, opts = {}) {
  const res = await fetch(url, opts);
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
  return data;
}

async function loadDir(path) {
  el.loading.style.display = 'block';
  el.emptyHint.style.display = 'none';
  el.errorMsg.style.display = 'none';
  el.fileList.innerHTML = '';
  try {
    const data = await api(`/api/files?path=${encodeURIComponent(path)}`);
    state.path = data.path || path;
    state.entries = data.entries || [];
    render();
  } catch (e) {
    showError(e.message);
  } finally {
    el.loading.style.display = 'none';
  }
}

// ── render ──────────────────────────────────────────────────────────
function render() {
  renderBreadcrumb();
  renderFileList();
}

function renderBreadcrumb() {
  const parts = state.path ? state.path.split('/') : [];
  let html = '<a data-path="">📁 FileHub</a>';
  let accum = '';
  for (const part of parts) {
    if (!part) continue;
    accum = accum ? `${accum}/${part}` : part;
    html += `<span class="sep">/</span><a data-path="${escAttr(accum)}">${escHtml(part)}</a>`;
  }
  el.breadcrumb.innerHTML = html;

  // Click handlers
  $$('a', el.breadcrumb).forEach(a => {
    a.addEventListener('click', () => {
      const p = a.dataset.path || '';
      navigate(p);
    });
  });
}

function renderFileList() {
  el.fileList.innerHTML = '';

  if (state.entries.length === 0) {
    el.emptyHint.style.display = 'block';
    return;
  }
  el.emptyHint.style.display = 'none';

  for (const entry of state.entries) {
    const li = document.createElement('li');
    const icon = entry.isDir ? '📁' : fileIcon(entry.name);
    const sizeStr = entry.isDir ? '' : formatSize(entry.size);

    li.innerHTML = `
      <span class="icon">${icon}</span>
      <span class="name">${escHtml(entry.name)}</span>
      <span class="meta">${entry.modTime} ${sizeStr}</span>
    `;
    li.dataset.name = entry.name;
    li.dataset.isDir = entry.isDir ? '1' : '0';
    li.dataset.preview = entry.preview ? '1' : '0';

    // Click: navigate into dir or download file
    li.addEventListener('click', () => {
      const targetPath = state.path ? `${state.path}/${entry.name}` : entry.name;
      if (entry.isDir) {
        navigate(targetPath);
      } else {
        // Download file
        window.open(`/api/files?path=${encodeURIComponent(targetPath)}&download=1`, '_blank');
      }
    });

    // Right-click: context menu
    li.addEventListener('contextmenu', (e) => {
      e.preventDefault();
      state.contextTarget = {
        name: entry.name,
        isDir: entry.isDir,
        preview: entry.preview,
        path: state.path ? `${state.path}/${entry.name}` : entry.name,
      };
      showContextMenu(e.clientX, e.clientY, entry);
    });

    el.fileList.appendChild(li);
  }
}

function fileIcon(name) {
  const ext = name.split('.').pop()?.toLowerCase();
  const map = {
    jpg: '🖼️', jpeg: '🖼️', png: '🖼️', gif: '🖼️', webp: '🖼️', svg: '🖼️', bmp: '🖼️', ico: '🖼️',
    pdf: '📄', doc: '📝', docx: '📝', xls: '📊', xlsx: '📊', ppt: '📽️', pptx: '📽️',
    zip: '📦', rar: '📦', '7z': '📦', tar: '📦', gz: '📦',
    mp4: '🎬', avi: '🎬', mkv: '🎬', mov: '🎬', webm: '🎬',
    mp3: '🎵', wav: '🎵', flac: '🎵', aac: '🎵', ogg: '🎵',
    txt: '📃', md: '📃', json: '📃', xml: '📃', yml: '📃', yaml: '📃',
    js: '💛', ts: '💙', jsx: '💛', tsx: '💙', py: '🐍', go: '🔵', rs: '🦀',
    html: '🌐', css: '🎨', scss: '🎨',
    exe: '⚙️', dmg: '💿', apk: '📱',
    sh: '💻', bat: '💻', ps1: '💻',
  };
  return map[ext] || '📄';
}

// ── navigation ──────────────────────────────────────────────────────
function navigate(path) {
  // Update URL hash
  window.location.hash = path ? `#/${path}` : '#';
  loadDir(path);
}

function escHtml(s) {
  const div = document.createElement('div');
  div.textContent = s;
  return div.innerHTML;
}

function escAttr(s) {
  return s.replace(/"/g, '&quot;').replace(/&/g, '&amp;');
}

function formatSize(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

// ── context menu ────────────────────────────────────────────────────
function showContextMenu(x, y, entry) {
  const menu = el.contextMenu;
  menu.classList.remove('hidden');
  menu.style.left = x + 'px';
  menu.style.top = y + 'px';

  // Show/hide preview button
  const previewBtn = $('[data-action="preview"]', menu);
  if (previewBtn) {
    previewBtn.style.display = (entry.preview && !entry.isDir) ? '' : 'none';
  }
}

function closeContextMenu() {
  el.contextMenu.classList.add('hidden');
  state.contextTarget = null;
}

document.addEventListener('click', (e) => {
  if (!el.contextMenu.contains(e.target)) {
    closeContextMenu();
  }
});

el.contextMenu.addEventListener('click', (e) => {
  const btn = e.target.closest('[data-action]');
  if (!btn) return;
  const action = btn.dataset.action;

  // Capture target BEFORE closing menu (closeContextMenu nullifies contextTarget)
  const target = state.contextTarget;
  state.pendingTarget = target;
  closeContextMenu();

  if (!target) return;

  switch (action) {
    case 'rename':
      openRenameDialog(target);
      break;
    case 'delete':
      openDeleteDialog(target);
      break;
    case 'preview':
      openPreview(target);
      break;
  }
});

// ── dialogs ─────────────────────────────────────────────────────────
function openDialog(dialog) {
  state.currentDialog = dialog;
  dialog.showModal();
}

function closeDialog() {
  if (state.currentDialog) {
    state.currentDialog.close();
    state.currentDialog = null;
  }
}

document.addEventListener('click', (e) => {
  if (e.target.dataset.close !== undefined) {
    closeDialog();
  }
});

// Close dialogs on backdrop click
document.addEventListener('click', (e) => {
  if (e.target.tagName === 'DIALOG' && state.currentDialog === e.target) {
    closeDialog();
  }
});

// ── new directory ───────────────────────────────────────────────────
$('#btn-newdir').addEventListener('click', () => {
  el.inputNewDir.value = '';
  openDialog(el.dialogNewDir);
  setTimeout(() => el.inputNewDir.focus(), 100);
});

el.dialogNewDir.querySelector('form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const name = el.inputNewDir.value.trim();
  if (!name) return;

  const dirPath = state.path ? `${state.path}/${name}` : name;
  try {
    await api('/api/dirs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path: dirPath }),
    });
    closeDialog();
    loadDir(state.path);
  } catch (e) {
    alert('创建失败: ' + e.message);
  }
});

// ── rename ──────────────────────────────────────────────────────────
function openRenameDialog(target) {
  el.inputRename.value = target.name;
  openDialog(el.dialogRename);
  setTimeout(() => {
    el.inputRename.focus();
    // Select name without extension for files
    if (!target.isDir) {
      const dot = target.name.lastIndexOf('.');
      if (dot > 0) {
        el.inputRename.setSelectionRange(0, dot);
        return;
      }
    }
    el.inputRename.select();
  }, 100);
}

el.dialogRename.querySelector('form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const newName = el.inputRename.value.trim();
  const target = state.pendingTarget;
  if (!newName || !target) return;

  try {
    await api('/api/rename', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path: target.path, newName }),
    });
    closeDialog();
    state.pendingTarget = null;
    loadDir(state.path);
  } catch (e) {
    alert('重命名失败: ' + e.message);
  }
});

// ── delete ──────────────────────────────────────────────────────────
function openDeleteDialog(target) {
  el.deleteName.textContent = `确定要删除「${target.name}」吗？${target.isDir ? '目录内所有内容将被删除。' : '此操作不可撤销。'}`;
  openDialog(el.dialogDelete);
}

el.dialogDelete.querySelector('form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const target = state.pendingTarget;
  if (!target) return;

  try {
    await api(`/api/files?path=${encodeURIComponent(target.path)}`, { method: 'DELETE' });
    closeDialog();
    state.pendingTarget = null;
    loadDir(state.path);
  } catch (e) {
    alert('删除失败: ' + e.message);
  }
});

// ── preview ─────────────────────────────────────────────────────────
function openPreview(target) {
  el.previewImg.src = `/api/preview?path=${encodeURIComponent(target.path)}`;
  el.previewName.textContent = target.name;
  openDialog(el.dialogPreview);
}

// Close preview on image click
$('#preview-close').addEventListener('click', () => closeDialog());

// ── upload ──────────────────────────────────────────────────────────
$('#btn-upload').addEventListener('click', () => {
  el.fileInput.click();
});

el.fileInput.addEventListener('change', () => {
  if (el.fileInput.files.length > 0) {
    uploadFiles(el.fileInput.files);
    el.fileInput.value = '';
  }
});

// Drag and drop
let dragCounter = 0;

el.dropZone.addEventListener('dragenter', (e) => {
  e.preventDefault();
  dragCounter++;
  el.dropZone.classList.add('drag-over');
});

el.dropZone.addEventListener('dragleave', () => {
  dragCounter--;
  if (dragCounter <= 0) {
    el.dropZone.classList.remove('drag-over');
    dragCounter = 0;
  }
});

el.dropZone.addEventListener('dragover', (e) => {
  e.preventDefault();
});

el.dropZone.addEventListener('drop', (e) => {
  e.preventDefault();
  dragCounter = 0;
  el.dropZone.classList.remove('drag-over');

  const files = e.dataTransfer.files;
  if (files.length > 0) {
    uploadFiles(files);
  }
});

async function uploadFiles(files) {
  const formData = new FormData();
  formData.append('path', state.path || '');
  for (const f of files) {
    formData.append('files', f);
  }

  try {
    el.loading.style.display = 'block';
    const res = await fetch('/api/files', { method: 'POST', body: formData });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Upload failed');
    loadDir(state.path);
  } catch (e) {
    alert('上传失败: ' + e.message);
    el.loading.style.display = 'none';
  }
}

// ── error ───────────────────────────────────────────────────────────
function showError(msg) {
  el.errorMsg.style.display = 'block';
  el.errorMsg.textContent = '⚠️ ' + msg;
}

// ── init ────────────────────────────────────────────────────────────
function init() {
  // Read path from URL hash
  const hash = window.location.hash;
  let initialPath = '';
  if (hash.startsWith('#/')) {
    initialPath = decodeURIComponent(hash.slice(2));
  }

  loadDir(initialPath);

  // Handle hash changes
  window.addEventListener('hashchange', () => {
    const h = window.location.hash;
    const p = h.startsWith('#/') ? decodeURIComponent(h.slice(2)) : '';
    loadDir(p);
  });

  // Keyboard shortcuts
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      closeContextMenu();
      closeDialog();
    }
    // Ctrl+Shift+N = new directory
    if (e.ctrlKey && e.shiftKey && e.key === 'N') {
      e.preventDefault();
      $('#btn-newdir').click();
    }
    // Ctrl+U = upload
    if (e.ctrlKey && !e.shiftKey && e.key === 'u') {
      e.preventDefault();
      $('#btn-upload').click();
    }
    // Delete key on selected file
    if (e.key === 'Delete') {
      const selected = el.fileList.querySelector('li.selected');
      if (selected) {
        const name = selected.dataset.name;
        const target = {
          name,
          isDir: selected.dataset.isDir === '1',
          preview: selected.dataset.preview === '1',
          path: state.path ? `${state.path}/${name}` : name,
        };
        state.pendingTarget = target;
        openDeleteDialog(target);
      }
    }
  });

  // Click to deselect
  el.fileList.addEventListener('click', (e) => {
    const li = e.target.closest('li');
    if (!li) return;
    // Toggle selected
    const wasSelected = li.classList.contains('selected');
    $$('li.selected', el.fileList).forEach(l => l.classList.remove('selected'));
    if (!wasSelected) {
      li.classList.add('selected');
    }
  });
}

init();
