<template>
  <div class="app">
    <!-- Title Bar -->
    <div class="title-bar">
      <div class="title">LinkRouter Config Editor</div>
      <div class="title-bar-buttons">
        <table>
          <tbody>
            <tr>
              <td><button class="titlebar-btn minimize-btn" @click="minimizeWindow">―</button></td>
              <td><button class="titlebar-btn maximize-btn" @click="maximizeWindow">◻</button></td>
              <td><button class="titlebar-btn close-btn" @click="closeWindow">⨯</button></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="header">
      <input
        v-model="search"
        placeholder="Search rules..."
        class="search-input"
      />
    </div>

    <!-- Scrollable Content Area -->
    <div class="content">
      <table class="config-table">
        <thead>
          <tr>
            <th style="width:3%; text-align: center;">#</th>
            <th style="width:62%">Regex</th>
            <th style="width:30%">Program</th>
            <th style="width:5%; min-width:32px;"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(item, idx) in filteredRules"
            :key="item.originalIndex"
            class="rule-row"
            :class="{ 'selected': selectedRowIndex === idx }"
            @dragstart="onDragStart($event, idx)"
            @dragover.prevent="onDragOver($event)"
            @drop="onDrop($event, idx)"
            @dblclick="openEditModal(item.rule)"
            @click="selectRow(idx)"
            @contextmenu.prevent="openContextMenu($event, item.rule, idx)"
          >
            <td class="index-cell">{{ item.originalIndex }}</td>
            <td>
              <div class="code-wrapper">
                <code>{{ item.rule.regex }}</code>
                <button
                  class="copy-btn"
                  @click.stop="copyToClipboard(item.rule.regex)"
                  title="Copy to clipboard"
                >
                  📋
                </button>
              </div>
            </td>
            <td>
              <div class="code-wrapper">
                <code>{{ basename(item.rule.program) }}</code>
                <button
                  class="copy-btn"
                  @click.stop="copyToClipboard(item.rule.program)"
                  title="Copy to clipboard"
                >
                  📋
                </button>
              </div>
            </td>
            <td class="drag-handle-cell">
              <div
                class="drag-handle"
                draggable="true"
                @dragstart="onDragStart($event, idx)"
              >≡</div>
            </td>
          </tr>
          <tr v-if="filteredRules.length === 0">
            <td colspan="4" style="text-align: center; padding: 2rem; color: #64748b;">
              No rules found<br>
              <small>Double-click a row to edit it</small>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Fixed Bottom Bar -->
    <div class="bottom-bar">
      <div class="status-info">
        {{ filteredRules.length }} of {{ config.rules?.length || 0 }} rules
        <!-- <span v-if="configPath" class="config-path">• {{ configPath }}</span> -->
        <span class="config-path" :class="{ 'saving': isSaving }">
          • {{ statusMessage || (configPath ? `${configPath}` : '') }}
        </span>
      </div>

      <div class="button-group">
        <button class="add-rule-btn" @click="openAddRuleModal">➕</button>
        <button class="save-btn" @click="saveConfigAs">💾</button>
        <button class="settings-btn" @click="openSettingsModal" title="Global Settings">⚙️</button>
        <button class="load-btn" @click="loadConfig">Load config</button>
      </div>
    </div>

    <!-- Edit Rule Modal -->
    <div v-if="showEditModal" class="modal-overlay"> <!--  @mousedown.self="closeEditModal" -->
      <div class="modal" @click.stop>
        <h2>Edit Rule</h2>

        <label>Pattern (Regex)</label>
        <input 
          v-model="editingRule.regex" 
          class="modal-input" 
          :class="{ 'invalid-regex': regexError }"
          @input="validateRegex"
          placeholder="e.g. ^https?://(.*\\.)?youtube\\.com/.*" 
        />
        <div v-if="regexError" class="regex-error-message">
          {{ regexError }}
        </div>

        <label>Program</label>
        <div class="program-input-wrapper">
          <input 
            v-model="editingRule.program" 
            class="modal-input program-input" 
            placeholder="C:\\Program Files\\App\\app.exe" 
          />
          <button class="browse-btn" @click="browseProgram" title="Browse for program">
            📂
          </button>
        </div>

        <label>Arguments (optional)</label>
        <input
          v-model="editingRule.arguments"
          class="modal-input"
          placeholder="--url {url}"
        />

        <label>Test URL</label>
        <div class="test-url-wrapper">
          <input
            v-model="testUrl"
            class="modal-input test-url-input"
            :class="{ 'match': testResult === true, 'no-match': testResult === false }"
            placeholder="URL to test regex"
            @input="updateTestResult"
          />
        </div>

        <div class="modal-buttons">
          <button class="cancel-btn" @click="closeEditModal">Cancel</button>
          <button class="ok-btn" @click="saveRule">Save</button>
        </div>
      </div>
    </div>

    <!-- Global Settings Modal -->
    <div v-if="showSettingsModal" class="modal-overlay">
      <div class="modal" @click.stop>
        <h2>Global Settings</h2>

        <label>Fallback Browser Path</label>
        <input
          v-model="editingGlobal.fallbackBrowserPath"
          class="modal-input"
          placeholder="e.g. C:\\Program Files\\Firefox\\firefox.exe"
        />

        <label>Fallback Browser Arguments</label>
        <input
          v-model="editingGlobal.fallbackBrowserArgs"
          class="modal-input"
          placeholder="e.g. -private-window {url}"
        />

        <label>
          Interactive Mode
          <input type="checkbox" v-model="editingGlobal.interactiveMode" />
        </label>

        <label>Default Config Editor</label>
        <input
          v-model="editingGlobal.defaultConfigEditor"
          class="modal-input"
          placeholder="e.g. notepad.exe"
        />

        <label>Log Path</label>
        <input
          v-model="editingGlobal.logPath"
          class="modal-input"
          placeholder="e.g. C:\\logs\\linkrouter.log"
        />

        <label>Supported Protocols (comma-separated)</label>
        <input
          v-model="protocolsInput"
          class="modal-input"
          placeholder="e.g. http,https,ftp"
        />

        <div class="modal-buttons">
          <button class="cancel-btn" @click="closeSettingsModal">Cancel</button>
          <button class="ok-btn" @click="okGlobalSettings">Save</button>
        </div>
      </div>
    </div>

    <!-- Context Menu -->
    <div
      v-if="contextMenu.visible"
      class="context-menu"
      :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }"
      @click.stop
    >
      <button class="context-item" @click="openAddRuleModal">➕ Add</button>
      <button class="context-item" @click="handleContextAction('edit')">✏️ Edit</button>
      <button class="context-item" @click="handleContextAction('delete')">🗑️ Delete</button>
    </div>

    <!-- Context Menu Backdrop -->
    <div
      v-if="contextMenu.visible"
      class="context-backdrop"
      @click="closeContextMenu"
      @contextmenu.prevent="closeContextMenu"
    ></div>
  </div>
</template>

<script setup>
import { Fzf } from 'fzf'
import { WindowMinimise, WindowToggleMaximise, Quit, LogInfo } from '../wailsjs/runtime/runtime';
import { ref, computed, nextTick } from 'vue';
import {
  GetInteractiveMode,
  OpenFileDialog,
  OpenConfigDialog,
  OpenProgramDialog,
  LoadConfigFromPath,
  SaveConfig,
  SaveConfigAs,
  GetConfig,
  GetCurrentConfigPath,
  TestRegex,
  IsValidRegex
} from '../wailsjs/go/main/App';

// Initial data load
async function loadInitialData() {
  try {
    const [cfg, path] = await Promise.all([
      GetConfig(),
      GetCurrentConfigPath()
    ]);
    config.value = cfg;
    configPath.value = path;
    saveToUndo();
  } catch (err) {
    console.error("Failed to load initial config/path:", err);
    config.value = { rules: [], global: {} };
    configPath.value = '';
  }
}

loadInitialData();

Promise.all([
  GetConfig(),
  GetCurrentConfigPath()
]).then(([cfg, path]) => {
  cfg.rules = (cfg.rules || []).map((rule, index) => ({
    ...rule,
    id: rule.id || `rule-${index}-${Date.now()}`
  }));
  config.value = cfg;
  configPath.value = path;
  saveToUndo();
});

const closeOnEsc = (e) => {
  if (e.key === 'Escape') {
    if (showEditModal.value) {
      closeEditModal()
    } else if (showSettingsModal.value) {
      closeSettingsModal()
    }
  }
}

nextTick(async () => {
  const mode = await GetInteractiveMode()
  if (mode.enabled === "true" && mode.url) {
    testUrl.value = mode.url
    editingRule.value = {
      regex: guessRegex(mode.url),
      program: '',
      arguments: ''
    }
    originalRule.value = null
    showEditModal.value = true
    updateTestResult()
  }
  window.addEventListener('keydown', closeOnEsc)
  window.addEventListener('keydown', (e) => {
    if (e.ctrlKey && e.key === 'z' && !e.shiftKey) {
      e.preventDefault()
      undo()
    }
  })
  window.addEventListener('keydown', (e) => {
    if (e.ctrlKey && e.key === 'y' && !e.shiftKey) {
      e.preventDefault()
      redo()
    }
  })
})

const guessRegex = (url) => {
  try {
    const u = new URL(url)
    return `${u.protocol}//${u.hostname}.*`
  } catch (e) {
    return '.*'
  }
}

// Reactive state
const config = ref({});
const configPath = ref('');
const search = ref('');

const showEditModal = ref(false);
const showSettingsModal = ref(false);
const selectedRowIndex = ref(-1);

const editingRule = ref({ regex: '', program: '', arguments: '' });
const originalRule = ref(null);

const editingGlobal = ref({
  fallbackBrowserPath: '',
  fallbackBrowserArgs: '',
  defaultConfigEditor: '',
  logPath: '',
  supportedProtocols: []
});
const originalGlobal = ref(null);
const protocolsInput = ref('');

const testUrl = ref('');
const testResult = ref(null);

const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  rule: null,
  index: -1
});

const dragSourceIndex = ref(-1);

// Fuzzy search
const fzfItems = computed(() => {
  return (config.value.rules || []).map((rule, i) => ({
    rule,
    realIndex: i,
    originalIndex: i + 1,
    str: `${rule.regex || ''} ${basename(rule.program || '')} ${rule.arguments || ''}`.toLowerCase()
  }))
})

const filteredRules = computed(() => {
  const query = search.value.trim()

  if (!query) {
    return (config.value.rules || []).map((rule, i) => ({
      rule,
      realIndex: i,
      originalIndex: i + 1
    }))
  }

  const fzf = new Fzf(fzfItems.value, {
    selector: (item) => item.str
  })

  const results = fzf.find(query)

  return results.map(entry => ({
    rule: entry.item.rule,
    realIndex: entry.item.realIndex,
    originalIndex: entry.item.originalIndex + 1,
    score: entry.score
  }))
})

// Window controls
const closeWindow = () => Quit();
const minimizeWindow = () => WindowMinimise();
const maximizeWindow = () => WindowToggleMaximise();

// Utility functions
function basename(path) {
  if (!path) return '';
  const parts = path.replace(/\\/g, '/').split('/');
  return parts[parts.length - 1] || path;
}

async function copyToClipboard(text) {
  if (!text) return;
  try {
    await navigator.clipboard.writeText(text);
  } catch (err) {
    console.error('Failed to copy:', err);
    const textarea = document.createElement('textarea');
    textarea.value = text;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    document.body.removeChild(textarea);
  }
}

// Config operations
const loadConfig = async () => {
  try {
    const filePath = await OpenConfigDialog();
    if (!filePath) return;

    const newConfig = await LoadConfigFromPath(filePath);
    if (newConfig) {
      config.value = newConfig;
      configPath.value = filePath;
      search.value = '';
    }
  } catch (err) {
    alert("Failed to load config: " + err);
    console.error(err);
  }
};

const saveConfigAs = async () => {
  try {
    const newPath = await SaveConfigAs(config.value);
    if (newPath) {
      configPath.value = newPath;
      showSavedNotification();
    }
  } catch (err) {
    alert('Failed to save config as: ' + err);
    console.error(err);
  }
};

const saveConfig = async () => {
  try {
    const Path = await SaveConfig(config.value);
    if (Path) {
      configPath.value = Path;
    }
  } catch (err) {
    alert('Failed to save config: ' + err);
    console.error(err);
  }
  showSavedNotification();
};

// Regex check
const regexError = ref('')

const validateRegex = async () => {
  const regexStr = editingRule.value.regex?.trim() || ''
  
  if (!regexStr) {
    regexError.value = ''
    return
  }

  const errMsg = await IsValidRegex(regexStr)
  regexError.value = errMsg

  updateTestResult()
}

// Rule editing
const openAddRuleModal = () => {
  editingRule.value = { regex: '', program: '', arguments: '' };
  originalRule.value = null;
  showEditModal.value = true;
  closeContextMenu();
};

const openEditModal = (rule) => {
  editingRule.value = {
    regex: rule.regex || '',
    program: rule.program || '',
    arguments: rule.arguments || ''
  };
  originalRule.value = rule;
  showEditModal.value = true;
  updateTestResult();
};

const closeEditModal = () => {
  showEditModal.value = false;
  setTimeout(() => {
    editingRule.value = { regex: '', program: '', arguments: '' };
    originalRule.value = null;
    regexError.value = null;
  }, 300);
};

const saveRule = () => {
  if (!editingRule.value.regex || !editingRule.value.program) {
    alert('Regex and Program are required!');
    return;
  }

  if (regexError.value) {
    alert('Please fix the regex syntax:\n' + regexError.value)
    return
  }

  if (originalRule.value) {
    Object.assign(originalRule.value, editingRule.value);
  } else {
    config.value.rules = config.value.rules || [];
    config.value.rules.push({ ...editingRule.value });
  }

  closeEditModal();
  saveConfig();
  saveToUndo();
};

const updateTestResult = async () => {
  const regex = editingRule.value?.regex?.trim() || '';
  const url = testUrl.value?.trim() || '';

  if (!regex || !url) {
    testResult.value = null;
    return;
  }

  try {
    const matches = await TestRegex(regex, url);
    testResult.value = matches;
  } catch (err) {
    testResult.value = false;
  }
};

const browseProgram = async () => {
  try {
    const filePath = await OpenProgramDialog()
    if (filePath) {
      editingRule.program = filePath
    }
  } catch (err) {
    console.error("Program picker failed:", err)
  }
}

// Global settings
const openSettingsModal = () => {
  editingGlobal.value = {
    fallbackBrowserPath: config.value.global?.fallbackBrowserPath || '',
    fallbackBrowserArgs: config.value.global?.fallbackBrowserArgs || '',
    defaultConfigEditor: config.value.global?.defaultConfigEditor || '',
    logPath: config.value.global?.logPath || '',
    interactiveMode: config.value.global?.interactiveMode || false,
    supportedProtocols: [...(config.value.global?.supportedProtocols || [])]
  };
  originalGlobal.value = config.value.global;
  protocolsInput.value = editingGlobal.value.supportedProtocols.join(', ');
  showSettingsModal.value = true;
};

const closeSettingsModal = () => {
  showSettingsModal.value = false;
  setTimeout(() => {
    editingGlobal.value = {
      fallbackBrowserPath: '',
      fallbackBrowserArgs: '',
      defaultConfigEditor: '',
      logPath: '',
      supportedProtocols: []
    };
    originalGlobal.value = null;
  }, 300);
};

const okGlobalSettings = async () => {
  try {
    editingGlobal.value.supportedProtocols = protocolsInput.value
      .split(',')
      .map(s => s.trim())
      .filter(s => s.length > 0);

    if (!config.value.global) {
      config.value.global = {};
    }

    Object.assign(config.value.global, editingGlobal.value);

    LogInfo('Protocols being saved: ' + JSON.stringify(editingGlobal.value.supportedProtocols));

    closeSettingsModal();
  } catch (err) {
    alert('Failed to save settings: ' + err);
    console.error(err);
  }
    saveConfig();
    saveToUndo();
};

// Row selection & context menu
function selectRow(index) {
  selectedRowIndex.value = index;
}

function openContextMenu(event, rule, index) {
  selectRow(index);

  const menuWidth = 180;
  const menuHeight = 120;

  let x = event.clientX;
  let y = event.clientY;

  if (x + menuWidth > window.innerWidth) {
    x = window.innerWidth - menuWidth - 2;
  }
  if (y + menuHeight > window.innerHeight) {
    y = window.innerHeight - menuHeight - 2;
  }

  x = Math.max(x, 2);
  y = Math.max(y, 2);

  contextMenu.value = {
    visible: true,
    x,
    y,
    rule,
    index
  };
}

function closeContextMenu() {
  contextMenu.value.visible = false;
}

function handleContextAction(action) {
  const { rule, index } = contextMenu.value;

  if (action === 'edit') {
    openEditModal(rule);
  } else if (action === 'delete') {
    if (confirm(`Delete rule:\n${rule.regex}\n→ ${rule.program}?`)) {
      const actualIndex = config.value.rules.findIndex(r =>
        r.regex === rule.regex &&
        r.program === rule.program &&
        r.arguments === rule.arguments
      );
      if (actualIndex !== -1) {
        config.value.rules.splice(actualIndex, 1);
      }
      if (selectedRowIndex.value === index) {
        selectedRowIndex.value = -1;
      }
      saveToUndo();
    }
  }

  closeContextMenu();
}

// Drag and drop
function onDragStart(event, index) {
  dragSourceIndex.value = index;
  event.dataTransfer.effectAllowed = 'move';
}

function onDragOver(event) {
  event.preventDefault();
  event.dataTransfer.dropEffect = 'move';

  document.querySelectorAll('.rule-row.drag-over').forEach(el => {
    el.classList.remove('drag-over');
  });
  const row = event.target.closest('.rule-row');
  if (row) row.classList.add('drag-over');
}

function onDrop(event, targetIndex) {
  event.preventDefault();

  const sourceIdx = dragSourceIndex.value;
  if (sourceIdx === -1 || sourceIdx === targetIndex) return;

  const rules = config.value.rules || [];
  if (rules.length === 0) return;

  const sourceRule = filteredRules.value[sourceIdx]?.rule;
  const targetRule = filteredRules.value[targetIndex]?.rule;

  if (!sourceRule || !targetRule) return;

  const realSource = rules.findIndex(r => r === sourceRule);
  const realTarget = rules.findIndex(r => r === targetRule);

  if (realSource === -1 || realTarget === -1 || realSource === realTarget) return;

  const updated = [...rules];
  const [moved] = updated.splice(realSource, 1);
  updated.splice(realTarget, 0, moved);

  config.value.rules = updated;

  selectedRowIndex.value = targetIndex;
  dragSourceIndex.value = -1;
  saveToUndo();
}


// config info
const statusMessage = ref('')
const configPathDisplay = computed(() => {
  return statusMessage.value || (configPath.value ? `${configPath.value}` : '')
})

const isSaving = ref(false)

const showSavedNotification = () => {
  isSaving.value = true
  statusMessage.value = 'Config saved!'

  setTimeout(() => {
    isSaving.value = false
      statusMessage.value = ''
  }, 2500)
}

// History Ctrl+z
const history = ref([])
const historyIndex = ref(-1) // -1 = no undo available

const saveToUndo = () => {
  // Only save if we're at the latest state
  if (historyIndex.value < history.value.length - 1) {
    // Discard future if we're in the middle
    history.value = history.value.slice(0, historyIndex.value + 1)
  }

  const deepCopy = JSON.parse(JSON.stringify(config.value))
  history.value.push(deepCopy)
  historyIndex.value = history.value.length - 1

  // Limit history size
  if (history.value.length > 30) {
    history.value.shift()
    historyIndex.value--
  }
}

// Undo
const undo = () => {
  if (historyIndex.value <= 0) return

  historyIndex.value--
  config.value = JSON.parse(JSON.stringify(history.value[historyIndex.value]))
  saveConfig();
}

// Redo
const redo = () => {
  if (historyIndex.value >= history.value.length) return

  historyIndex.value++
  config.value = JSON.parse(JSON.stringify(history.value[historyIndex.value]))
  saveConfig();
}
</script>
