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
      <div class="search-wrapper">
        <input
          ref="searchInput"
          v-model="search"
          placeholder="Search rules..."
          class="search-input"
        />
        <button 
          v-if="search" 
          class="search-clear" 
          @click="search = ''"
          title="Clear search"
        >
          ✕
        </button>
      </div>
    </div>

    <!-- Scrollable Content Area -->
    <div
      class="content" 
      ref="rulesContainer"
      tabindex="-1"
      @keydown="handleRulesKeydown"
    >
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
                  <span class="emoji">📋︎</span>
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
                  <span class="emoji">📋︎</span>
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

    <!-- Bottom Bar -->
    <div class="bottom-bar">
      <div class="status-info">
        {{ filteredRules.length }} of {{ config.rules?.length || 0 }} rules
        <span class="config-path" :class="{ 'saving': isSaving }" :key="notificationKey">
          • {{ statusMessage || (configPath ? `${configPath}` : '') }}
        </span>
      </div>

      <div class="button-group">
        <button class="add-rule-btn" @click="openAddRuleModal" title="Add new rule"><span class="emoji">➕︎</span></button>
        <button class="save-btn" @click="saveConfigAs" title="Save as"><span class="emoji">💾︎</span></button>
        <button class="settings-btn" @click="openSettingsModal" title="Global settings"><span class="emoji">⚙</span></button>
        <button class="load-btn" @click="loadConfig" title="Open config">📂︎</button>
      </div>
    </div>

    <!-- Edit Rule Modal -->
    <div v-if="showEditModal" class="modal-overlay"> <!--  @mousedown.self="closeEditModal" -->
      <div class="modal" @click.stop>
        <h2>Edit Rule</h2>

        <label>Pattern (Regex)</label>
        <input 
          ref="regexInput"
          v-model="editingRule.regex" 
          class="modal-input" 
          :class="{ 'invalid-regex': regexError }"
          @input="validateRegex"
          placeholder="e.g. ^https?://(.*\.)?youtube\.com/.*" 
        />
        <div v-if="regexError" class="regex-error-message">
          {{ regexError }}
        </div>

        <label>Program</label>
        <div class="program-input-wrapper">
          <input 
            v-model="editingRule.program" 
            class="modal-input program-input" 
            placeholder="C:\Program Files\App\app.exe" 
          />
          <button class="browse-btn" @click="browseFile('ruleProgram')" title="Browse for program">
            <span class="emoji">📂︎</span>
          </button>
        </div>

        <label>Arguments (optional)</label>
        <input
          v-model="editingRule.arguments"
          class="modal-input"
          placeholder="{URL} for URL; $1, $2 etc for captured groups"
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
        <div  style="text-align: left; margin-top: 0.5rem;">
          <button
            v-if="testUrl"
            class="browser-btn"
            @click="openTestUrlInBrowser"
            title="Open test URL in default browser"
          >
            🌐︎ Open in Browser
          </button>
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

        <div class="modal-form-content">

          <label>Fallback Browser Path</label>
          <div class="program-input-wrapper">
            <input
              ref="fallbackBrowserInput"
              v-model="editingGlobal.fallbackBrowserPath"
              class="modal-input"
              placeholder="e.g. C:\\Program Files\\Firefox\\firefox.exe"
            />
            <button class="browse-btn" @click="browseFile('fallbackBrowser')" title="Browse for program">
              <span class="emoji">📂︎</span>
            </button>
          </div>

          <label>Fallback Browser Arguments</label>
          <input
            v-model="editingGlobal.fallbackBrowserArgs"
            class="modal-input"
            placeholder="e.g. -private-window {url}"
          />

          <label>
            Interactive Mode
          </label>
          <input type="checkbox" v-model="editingGlobal.interactiveMode" />

          <label>Default Config Editor</label>
          <div class="program-input-wrapper">
            <input
            v-model="editingGlobal.defaultConfigEditor"
            class="modal-input"
            placeholder="e.g. notepad.exe"
            />
            <button class="browse-btn" @click="browseFile('defaultEditor')" title="Browse for program">
              <span class="emoji">📂︎</span>
            </button>
          </div>
          
          <label>Log Path</label>
          <div class="program-input-wrapper">
            <input
            v-model="editingGlobal.logPath"
            class="modal-input"
            placeholder="e.g. C:\\logs\\linkrouter.log"
            />
            <button class="browse-btn" @click="browseFile('logPath')" title="Browse for program">
              <span class="emoji">📂︎</span>
            </button>
          </div>
          
          <label>Supported Protocols (comma-separated)</label>
          <input
          v-model="protocolsInput"
          class="modal-input"
          placeholder="e.g. http,https,ftp"
          />
        </div>

        
        <div class="modal-button-bar">
          <div>
            <button class="reg-btn" @click="registerApp">Register</button>
            <button class="reg-btn warning" @click="unregisterApp">Unregister</button>
          </div>
          <div>
            <button class="cancel-btn" @click="closeSettingsModal">Cancel</button>
            <button class="ok-btn" @click="saveGlobalSettings">Save</button>
          </div>
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
      <button class="context-item" @click="openAddRuleModal">➕︎ &nbspAdd</button>
      <button class="context-item" @click="handleContextAction('edit')">✎ &nbspEdit</button>
      <button class="context-item" @click="handleContextAction('delete')">🗑 &nbsp&nbspDelete</button>
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
import { WindowMinimise, WindowToggleMaximise, Quit, LogInfo, EventsOn, WindowShow, WindowSetAlwaysOnTop, WindowUnminimise} from '../wailsjs/runtime/runtime';
import { ref, computed, nextTick, onMounted } from 'vue';
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
  IsValidRegex,
  RegisterLinkRouter,
  UnregisterLinkRouter,
  OpenInFallbackBrowser
} from '../wailsjs/go/main/App';

Promise.all([
  GetConfig(),
  GetCurrentConfigPath()
]).then(async ([cfg, path]) => {
  cfg.rules = (cfg.rules || []).map((rule, index) => ({
    ...rule,
    id: rule.id || `rule-${index}-${Date.now()}`
  }));
  config.value = cfg;
  configPath.value = path;
  saveToUndo();
  nextTick(() => {
    searchInput.value?.focus()
  });


  const mode = await GetInteractiveMode()
  launchedInInteractiveMode.value = mode.enabled === "true";
  if (mode.enabled === "true" && mode.url) {
    testUrl.value = mode.url
    editingRule.value = {
      regex: guessRegex(mode.url),
      program: '',
      arguments: ''
    }
    originalRule.value = null
    showEditModal.value = true
    editingRule.value.regex
    updateTestResult()
    nextTick(() => {
      regexInput.value?.focus()
    })
  }
  runtime.WindowMinimise()
  setTimeout(() => {
    runtime.WindowUnminimise()
  }, 20);
});

const guessRegex = (url) => {
  try {
    const u = new URL(url)
    return `${u.protocol}//${u.hostname}.*`
  } catch (e) {
    return '.*'
  }
}

// Keyboard Hotkeys
setTimeout(() => {
  window.addEventListener('keydown', (e) => {
    if (e.ctrlKey && e.key === 'Enter' && !e.shiftKey) {
      if (showSettingsModal.value || showEditModal.value) {
        e.preventDefault()
        saveGlobalSettings()
        saveRule()
        e.stopPropagation()
        return
      }
    }
    if (e.key === 'Enter') {
      if (document.activeElement !== rulesContainer.value && !showSettingsModal.value && !showEditModal.value) {
        e.preventDefault()
        rulesContainer.value?.focus()
        e.stopPropagation()
        return
      }
    }
    if (e.key === 'Escape') {
      if (showEditModal.value ) {
        e.preventDefault()
        closeEditModal()
        e.stopPropagation()
        return
      } else if (showSettingsModal.value) {
        e.preventDefault()
        closeSettingsModal()
        e.stopPropagation()
        return
      }
      if (contextMenu.value.visible) {
        e.preventDefault()
        closeContextMenu()
        e.stopPropagation()
        return
      }
      if (document.activeElement !== rulesContainer.value) {
        e.preventDefault()
        rulesContainer.value?.focus()
        e.stopPropagation()
        return
      }
      if (document.activeElement == rulesContainer.value) {
        e.preventDefault()
        searchInput.value?.focus()
        e.stopPropagation()
        return
      }
    }
    if (e.ctrlKey && e.key === 'z' && !e.shiftKey) {
      e.preventDefault()
      undo()
      e.stopPropagation()
      return
    }
    if (e.ctrlKey && e.key === 'y' && !e.shiftKey) {
      e.preventDefault()
      redo()
      e.stopPropagation()
      return
    }
    if (e.ctrlKey && e.key === 's' && !e.shiftKey) {
      if (showEditModal.value) {
        e.preventDefault()
        saveRule();
        e.stopPropagation()
        return
      } else if (showSettingsModal.value) {
        e.preventDefault()
        saveGlobalSettings()
        e.stopPropagation()
        return
      }
      e.preventDefault()
      saveConfigAs()
      e.stopPropagation()
      return
    }
    if ((e.ctrlKey && e.key === 'l' && !e.shiftKey) || (e.ctrlKey && e.key === 'f' && !e.shiftKey)) {
      e.preventDefault()
      searchInput.value?.focus()
      e.stopPropagation()
      return
    }
    if (e.key === '/' && !e.shiftKey && !e.altKey && !e.ctrlKey && !e.metaKey) {
      e.preventDefault()
      searchInput.value?.focus()
      e.stopPropagation()
      return
    }
    if (e.ctrlKey && e.key === 'n' && !e.shiftKey) {
      e.preventDefault()
      openAddRuleModal()
      e.stopPropagation()
      return
    }
    if (e.ctrlKey && e.key === 'o' && !e.shiftKey) {
      if (launchedInInteractiveMode.value && showEditModal.value) {
        e.preventDefault()
        openTestUrlInBrowser()
        e.stopPropagation()
        return
      }
    }
  })
}, 100)

// Reactive state
const config = ref({});
const configPath = ref('');
const search = ref('');
const rulesContainer = ref(null);
const searchInput = ref(null);
const fallbackBrowserInput = ref(null);
const launchedInInteractiveMode = ref(false);

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

const regexInput = ref(null);

const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  rule: null,
  index: -1
});

const dragSourceIndex = ref(-1);

// Fuzzy search

// const fzfItems = computed(() => {
//   return (config.value.rules || []).map((rule, i) => ({
//     rule,
//     realIndex: i,
//     originalIndex: i + 1,
//     str: `${rule.regex || ''} ${basename(rule.program || '')} ${rule.arguments || ''}`.toLowerCase()
//   }))
// })

// const filteredRules = computed(() => {
//   const query = search.value.trim()

//   if (!query) {
//     return (config.value.rules || []).map((rule, i) => ({
//       rule,
//       realIndex: i,
//       originalIndex: i + 1
//     }))
//   }

//   const fzf = new Fzf(fzfItems.value, {
//     selector: (item) => item.str
//   })

//   const matches = fzf.find(query)

//   const matchedRules = new Set(matches.map(entry => entry.item.rule))

//   const result = (config.value.rules || [])
//     .map((rule, i) => {
//       if (matchedRules.has(rule)) {
//         return {
//           rule,
//           realIndex: i,
//           originalIndex: i + 1
//         }
//       }
//       return null
//     })
//     .filter(Boolean)

//   return result
// })

const filteredRules = computed(() => {
  const query = search.value.trim()
  
  if (!query) {
    return (config.value.rules || []).map((rule, i) => ({
      rule,
      realIndex: i,
      originalIndex: i + 1
    }))
  }

  const terms = query.toLowerCase().split(/\s+/).filter(t => t)
  const rulesWithStr = (config.value.rules || []).map((rule, i) => ({
    rule,
    realIndex: i,
    originalIndex: i + 1,
    str: `${rule.regex || ''} ${basename(rule.program || '')} ${rule.arguments || ''}`.toLowerCase()
  }))


  const substringMatches = new Set()
  const regexMatches = new Set()

  for (const item of rulesWithStr) {
    if (terms.every(term => item.str.includes(term))) {
      substringMatches.add(item.rule)
    }
  }

  for (const item of rulesWithStr) {
    if (!item.rule.regex) continue
    try {
      const re = new RegExp(item.rule.regex, 'i')
      if (re.test(query)) {
        regexMatches.add(item.rule)
      }
    } catch (e) {
    }
  }

  const combinedMatches = new Set([...substringMatches, ...regexMatches])

  return rulesWithStr.filter(item => combinedMatches.has(item.rule))
})

const handleRulesKeydown = (e) => {
  const rules = config.value.rules || []
  if (rules.length === 0) return

  // Handle arrow keys
  if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
    e.preventDefault()

    const direction = e.key === 'ArrowUp' ? -1 : 1
    let newIndex = selectedRowIndex.value + direction

    // Wrap around or clamp? We'll clamp.
    if (newIndex < 0) newIndex = 0
    if (newIndex >= filteredRules.value.length) newIndex = filteredRules.value.length - 1

    selectedRowIndex.value = newIndex

    // Optional: auto-scroll to ensure selected row is visible
    nextTick(() => {
      const container = rulesContainer.value
      const selectedRow = container.querySelector('.rule-row.selected')
      if (selectedRow) {
        selectedRow.scrollIntoView({ block: 'nearest' })
      }
    })
  }

  // Handle Enter to edit
  if (e.key === 'Enter') {
    const index = selectedRowIndex.value
    if (index >= 0 && index < filteredRules.value.length) {
      const rule = filteredRules.value[index].rule
      openEditModal(rule)
    }
  }
}

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
  nextTick(() => {
    regexInput.value?.focus()
  });
};

const openEditModal = (rule) => {
  editingRule.value = {
    regex: rule.regex || '',
    program: rule.program || '',
    arguments: rule.arguments || ''
  };
  originalRule.value = rule;
  showEditModal.value = true;
  nextTick(() => {
    regexInput.value?.focus()
  });
  updateTestResult();
};

const closeEditModal = () => {
  showEditModal.value = false;
  setTimeout(() => {
    editingRule.value = { regex: '', program: '', arguments: '' };
    originalRule.value = null;
    regexError.value = null;
  }, 300);
  if (launchedInInteractiveMode.value) {
    runtime.Quit()
  }
};

const openTestUrlInBrowser = async () => {
  if (!testUrl.value?.trim()) return;
  try {
    await OpenInFallbackBrowser(config.value.global.fallbackBrowserPath, testUrl.value.trim());
  } catch (err) {
    runtime.LogError("Failed to open URL:", err);
  }
  if (launchedInInteractiveMode.value) {
    runtime.Quit()
  }
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

  saveConfig();
  closeEditModal();
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

const browseFile = async (type) => {
  try {
    if ( type == 'fallbackBrowser' ) {
      const filePath = await OpenProgramDialog()
      editingGlobal.value.fallbackBrowserPath = filePath
    }
    if ( type == 'defaultEditor' ) {
      const filePath = await OpenProgramDialog()
      editingGlobal.value.defaultConfigEditor = filePath
    }
    if ( type == 'logPath' ) {
      const filePath = await OpenFileDialog("Select Log File", [])
      editingGlobal.value.logPath = filePath
    }
    if ( type == 'ruleProgram' ) {
      const filePath = await OpenProgramDialog()
      editingRule.value.program = filePath
    }
  } catch (err) {
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
  // nextTick(() => {
  //   fallbackBrowserInput.value?.focus()
  // });
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

const saveGlobalSettings = async () => {
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
  }
    saveConfig();
    saveToUndo();
};

const registerApp = async () => {
    saveGlobalSettings()
    await RegisterLinkRouter()
}

const unregisterApp = async () => {
  if (!confirm('Unregister LinkRouter from the system?')) return
  try {
    await UnregisterLinkRouter()
  } catch (err) {
  }
}

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
      saveConfig();
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
  saveConfig();
  saveToUndo();
}


// config info
const statusMessage = ref('')
const configPathDisplay = computed(() => {
  return statusMessage.value || (configPath.value ? `${configPath.value}` : '')
})

const isSaving = ref(false)

let saveNotificationTimeout = null
const notificationKey = ref(0)

const showSavedNotification = () => {
  isSaving.value = true
  statusMessage.value = 'Config saved!'
  notificationKey.value++

  // Clear previous timer
  if (saveNotificationTimeout) {
    clearTimeout(saveNotificationTimeout)
  }

  // Set new timer from last call
  saveNotificationTimeout = setTimeout(() => {
    isSaving.value = false
    statusMessage.value = ''
  }, 4000)
}

// History Ctrl+z
const history = ref([])
const historyIndex = ref(-1)

const saveToUndo = () => {
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

const undo = () => {
  if (historyIndex.value <= 0) return

  historyIndex.value--
  config.value = JSON.parse(JSON.stringify(history.value[historyIndex.value]))
  saveConfig();
}

const redo = () => {
  if (historyIndex.value >= history.value.length) return

  historyIndex.value++
  config.value = JSON.parse(JSON.stringify(history.value[historyIndex.value]))
  saveConfig();
}
</script>
