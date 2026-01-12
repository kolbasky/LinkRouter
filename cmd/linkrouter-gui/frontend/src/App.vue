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
                <code>{{ !(copiedIndex === idx && copiedField === 'regex') ? item.rule.regex : 'Copied!'}}</code>
                <button
                  class="copy-btn"
                  @click.stop="copyToClipboard(item.rule.regex, idx, 'regex')"
                  :title="copiedIndex !== idx || copiedField !== 'regex' ? 'Copy to clipboard' : 'Copied!'"
                >
                  <span class="emoji" v-if="!(copiedIndex === idx && copiedField === 'regex')">📋︎</span>
                  <span class="emoji checkmark" v-else>✔</span>
                </button>
              </div>
            </td>
            <td>
              <div class="code-wrapper">
                <code>{{ !(copiedIndex === idx && copiedField === 'program') ? basename(item.rule.program) : 'Copied!'}}</code>
                <button
                  class="copy-btn"
                  @click.stop="copyToClipboard(item.rule.program, idx, 'program')"
                  :title="copiedIndex !== idx || copiedField !== 'program' ? 'Copy to clipboard' : 'Copied!'"
                >
                  <span class="emoji" v-if="!(copiedIndex === idx && copiedField === 'program')">📋︎</span>
                  <span class="emoji checkmark" v-else>✓</span>
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
            <td colspan="4" style="text-align: center; padding: 2rem; color: #64748b;" @click="openAddRuleModal">
              No rules found<br>
              <small>Click here to create new rule</small>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Bottom Bar -->
    <div class="bottom-bar">
      <div class="status-info">
        {{ filteredRules.length }} of {{ config.rules?.length || 0 }} rules
        <span class="config-path-wrapper">
          <span 
            v-if="configPath"
            class="config-path" 
            :class="{ 'report-success': reportSuccess }" 
            :key="notificationKey"
            @click="openConfigInExplorer"
            title="Open config path"
          >
            • {{ statusMessage || (configPath ? `${configPath}` : '') }}
          </span>
        </span>
      </div>

      <div class="button-group">
        <button class="add-rule-btn" @click="openAddRuleModal" title="Add new rule"><span class="emoji">➕&#65038</span></button>
        <button class="reload-btn" @click="reloadConfig(false)" title="Reload config from disk">🔄&#65038</button>
        <button class="load-btn" @click="loadConfig" title="Open config">📂︎&#65038</button>
        <button class="save-btn" @click="saveConfigAs" title="Save as"><span class="emoji">💾&#65038</span></button>
        <!-- <button class="settings-btn" @click="openSettingsModal" title="Global settings"><span class="emoji">🔧&#65038</span></button> -->
        <button class="settings-btn" @click="openSettingsModal" title="Global settings"><span class="emoji">⛭&#65038</span></button>
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
            🌐︎&nbsp&nbspOpen in browser
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
              placeholder="e.g. C:\Program Files\Firefox\firefox.exe"
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
            placeholder="e.g. C:\logs\linkrouter.log"
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
      <button class="context-item" @click="openAddRuleModal">➕︎&nbsp&nbspAdd</button><br>
      <button class="context-item" @click="handleContextAction('edit')">✎&nbsp&nbspEdit</button><br>
      <button class="context-item" @click="handleContextAction('copy')">⧉&nbsp&nbspCopy</button><br>
      <button class="context-item" :disabled="!clipboardRule" @click="handleContextAction('paste')">📋︎&nbsp&nbspPaste</button><br>
      <button class="context-item" @click="handleContextAction('duplicate')">⮻&nbsp&nbspDuplicate</button><br>
      <button class="context-item" @click="handleContextAction('delete')">❌︎&nbsp&nbspDelete</button><br>
    </div>

    <!-- Context Menu Backdrop -->
    <div
      v-if="contextMenu.visible"
      class="context-backdrop"
      @click="closeContextMenu"
      @contextmenu.prevent="closeContextMenu"
    ></div>
  </div>

  <!-- Alert Modal -->
  <div v-if="showAlert" class="modal-overlay" @click="showAlert = false">
    <div class="modal alert-modal" @click.stop>
      <p>{{ alertMessage }}</p>
      <div class="modal-buttons">
        <button class="ok-btn" @click="showAlert = false">OK</button>
      </div>
    </div>
  </div>

  <!-- Confirmation Modal -->
  <div v-if="showConfirm" class="modal-overlay" @click="showConfirm = false">
    <div class="modal confirm-modal" @click.stop>
      <h2>{{ confirmHeader }}</h2>
      <p>{{ confirmMessage }}</p>
      <div class="modal-buttons">
        <button class="cancel-btn" @click="showConfirm = false">{{ confirmCancelBtn }}</button>
        <button class="ok-btn" @click="handleConfirm">{{ confirmOkBtn }}</button>
      </div>
    </div>
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
    }, 100);
}).catch((err) => {
  showAlertModal(`Loading config failed:\n\n${err.message || err}`)
});

nextTick(() => {
  let lastFocus = 0;
  window.addEventListener('focus', () => {
    if (Date.now() - lastFocus > 3000) reloadConfig(true);
    lastFocus = Date.now();
  });
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
const isAnyModalOpen = computed(() => {
  return (
    showEditModal.value ||
    showSettingsModal.value ||
    showAlert.value ||
    showConfirm.value ||
    contextMenu.value.visible
  )
})

const isInputFocused = () => {
  const active = document.activeElement
  return (
    active?.tagName === 'INPUT' ||
    active?.tagName === 'TEXTAREA' ||
    active?.isContentEditable
  )
}

const shouldAllowGlobalShortcuts = () => {
  return !isAnyModalOpen.value && !isInputFocused()
}



setTimeout(() => {
  window.addEventListener('keydown', (e) => {
    const isCtrl  = e.ctrlKey || e.metaKey
    const isPlain = !e.shiftKey && !e.altKey && !e.metaKey && !e.ctrlKey

    // ─────────────────────────────────────────────────────────────
    // Always-working keys
    // ─────────────────────────────────────────────────────────────

    if (e.key === 'Escape') {
      e.preventDefault()

      if (showAlert.value)          { showAlert.value = false; return }
      if (showConfirm.value)        { showConfirm.value = false; return }
      if (showEditModal.value)      { closeEditModal(); return }
      if (showSettingsModal.value)  { closeSettingsModal(); return }
      if (contextMenu.value.visible){ closeContextMenu(); return }

      if (document.activeElement !== searchInput.value) {
        searchInput.value?.focus()
      } else {
        rulesContainer.value?.focus()
      }
      return
    }

    // ─────────────────────────────────────────────────────────────
    // Priority 2: Enter in modals (confirm / default action)
    // ─────────────────────────────────────────────────────────────

    if (e.key === 'Enter' && !isCtrl) {
      if (showAlert.value) {
        e.preventDefault()
        showAlert.value = false
        return
      }
      if (showConfirm.value) {
        e.preventDefault()
        handleConfirm()
        return
      }
      if (document.activeElement === searchInput.value) {
        e.preventDefault()
        rulesContainer.value?.focus()
        return
      }
    }

    if ((e.key === 'Enter' && isCtrl) && (showEditModal.value || showSettingsModal.value)) {
      e.preventDefault()
      if (showEditModal.value)  saveRule()
      if (showSettingsModal.value) saveGlobalSettings()
      return
    }
    
    // Test URL in browser (only when edit modal open)
    if (isCtrl && e.code === 'KeyO' && showEditModal.value) {
      e.preventDefault()
      openTestUrlInBrowser()
      return
    }

    // Save
    if (isCtrl && e.code === 'KeyS') {
      e.preventDefault()
      if (showEditModal.value)      { saveRule(); return }
      if (showSettingsModal.value)  { saveGlobalSettings(); return }
      saveConfig()
      return
    }

    if (!isAnyModalOpen.value) {
      if (isCtrl && e.code === 'KeyN') {
        e.preventDefault()
        openAddRuleModal()
        return
      }
    }
    // ─────────────────────────────────────────────────────────────
    // No modals opened, no inputs focused
    // ─────────────────────────────────────────────────────────────

    if (!shouldAllowGlobalShortcuts()) return

    // Undo / Redo
    if (isCtrl && !e.shiftKey) {
      if (e.code === 'KeyZ') { e.preventDefault(); undo(); return }
      if (e.code === 'KeyY') { e.preventDefault(); redo(); return }
    }

    // Search
    if ((isCtrl && (e.code === 'KeyF' || e.code === 'KeyL')) || (e.key === '/' && isPlain)) {
      e.preventDefault()
      searchInput.value?.focus()
      return
    }
  });
}, 100);

// Reactive state
const config = ref({});
const configPath = ref('');
const copiedIndex = ref(-1)
const copiedField = ref(null)
const search = ref('');
const rulesContainer = ref(null);
const searchInput = ref(null);
const fallbackBrowserInput = ref(null);
const launchedInInteractiveMode = ref(false);

const alertMessage = ref('');
const showAlert = ref(false);
const confirmHeader = ref('');
const confirmMessage = ref('');
const confirmOkBtn = ref('');
const confirmCancelBtn = ref('');
const onConfirm = ref(null); // callback function
const showConfirm = ref(false);

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

    // Clamp index
    newIndex = Math.max(0, Math.min(newIndex, filteredRules.value.length - 1))
    selectedRowIndex.value = newIndex

    nextTick(() => {
      const container = rulesContainer.value
      if (!container) return

      const selectedRow = container.querySelector('.rule-row.selected')
      if (!selectedRow) return

      // Option 1: Most reliable — scroll with exact offset to clear header
      const headerHeight = 60;
      const rowTop = selectedRow.getBoundingClientRect().top
      const containerTop = container.getBoundingClientRect().top

      if (rowTop < containerTop + headerHeight) {
        // Row is hidden under header → scroll it down just enough
        container.scrollTop += (rowTop - containerTop - headerHeight - 10) // +10px margin
      } else if (rowTop + selectedRow.offsetHeight > container.clientHeight + containerTop) {
        // Row is below viewport → scroll up
        container.scrollTop += (rowTop + selectedRow.offsetHeight - container.clientHeight - containerTop + 10)
      }
    })
  }

  // Handle Enter to edit
  if (e.key === 'Enter' && shouldAllowGlobalShortcuts()) {
    const index = selectedRowIndex.value
    if (index >= 0 && index < filteredRules.value.length) {
      const rule = filteredRules.value[index].rule
      openEditModal(rule)
    }
  }
  if ((e.key === 'Delete' || e.key === 'Backspace')) {
    const index = selectedRowIndex.value
    if (index >= 0 && index < filteredRules.value.length) {
      e.preventDefault();
      handleContextAction('delete', index);
      e.stopPropagation();
      return;
    }
  }

  const isCtrl = e.ctrlKey || e.metaKey;
  const isPlain = !e.shiftKey && !e.altKey && !e.metaKey && !e.ctrlKey;
  // Copy (Ctrl+C)
  if (isCtrl && e.code === 'KeyC' && selectedRowIndex.value >= 0) {
    e.preventDefault();
    handleContextAction('copy', selectedRowIndex.value);
    return;
  }

  // Paste (Ctrl+V)
  if (isCtrl && e.code === 'KeyV' && selectedRowIndex.value >= 0) {
    e.preventDefault();
    handleContextAction('paste', selectedRowIndex.value);
    return;
  }

  // Duplicate (Ctrl+D)
  if (isCtrl && e.code === 'KeyD' && selectedRowIndex.value >= 0) {
    e.preventDefault();
    handleContextAction('duplicate', selectedRowIndex.value);
    return;
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

async function copyToClipboard(text, rowIndex, field) {
  if (!text) return;

  try {
    await navigator.clipboard.writeText(text);
    
    copiedIndex.value = rowIndex;
    copiedField.value = field;
    
    setTimeout(() => {
      if (copiedIndex.value === rowIndex && copiedField.value === field) {
        copiedIndex.value = -1;
      }
    }, 2000);

  } catch (err) {
    // Fallback
    const textarea = document.createElement('textarea');
    textarea.value = text;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    document.body.removeChild(textarea);

    // Still show feedback on fallback
    copiedIndex.value = rowIndex;
    setTimeout(() => {
      if (copiedIndex.value === rowIndex && copiedField.value === field) copiedIndex.value = -1;
    }, 2000);
  }
}

// Config operations
const loadConfig = async () => {
  searchInput.value?.focus()
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
    showAlertModal(`Failed to load config:\n\n${err.message || err}`);
  }
};

const reloadConfig = async (silent = false) => {
  try {
    // Re-fetch config from backend (reads file again)
    const [cfg, path] = await Promise.all([
      GetConfig(),
      GetCurrentConfigPath()
    ]);

    // Preserve selection if possible
    const oldSelectedRule = selectedRowIndex.value >= 0 
      ? filteredRules.value[selectedRowIndex.value]?.rule 
      : null;

    // Update config
    cfg.rules = (cfg.rules || []).map((rule, index) => ({
      ...rule,
      id: rule.id || `rule-${index}-${Date.now()}`
    }));
    config.value = cfg;
    configPath.value = path;
    saveToUndo();

    // Try to restore selection
    if (oldSelectedRule) {
      nextTick(() => {
        const newIndex = filteredRules.value.findIndex(
          item => 
            item.rule.regex === oldSelectedRule.regex &&
            item.rule.program === oldSelectedRule.program &&
            item.rule.arguments === oldSelectedRule.arguments
        );
        if (newIndex !== -1) {
          selectedRowIndex.value = newIndex;
        }
      });
    }
    if (!silent) {
      showSavedNotification('Config reloaded!');
    }
  } catch (err) {
    showAlertModal(`Reload failed ${err.message || err}`);
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
    showAlertModal(`Failed to save config as:\n\n${err.message || err}`);
  }
  searchInput.value?.focus()
};

const saveConfig = async () => {
  try {
    const Path = await SaveConfig(config.value);
    if (Path) {
      configPath.value = Path;
    }
  } catch (err) {
    showAlertModal(`Failed to save config:\n\n${err.message || err}`);
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
  rulesContainer.value?.focus()
  if (launchedInInteractiveMode.value) {
    runtime.Quit()
  }
};

const openTestUrlInBrowser = async () => {
  if (!testUrl.value?.trim()) return;
  try {
    await OpenInFallbackBrowser(config.value.global.fallbackBrowserPath, '"' + testUrl.value.trim() + '"');
  } catch (err) {
    runtime.LogError("Failed to open URL:", err);
  }
  if (launchedInInteractiveMode.value) {
    runtime.Quit()
  }
};

const saveRule = () => {
  if (!editingRule.value.regex || !editingRule.value.program) {
    showAlertModal('Regex and Program are required!');
    return;
  }

  if (regexError.value) {
    showAlertModal('Please fix the regex syntax:\n\n' + regexError.value)
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
  searchInput.value?.focus()
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

    closeSettingsModal();
  } catch (err) {
    showAlertModal(`Failed to save settings:\n\n${err.message || err}`);
  }
    saveConfig();
    saveToUndo();
};

const registerApp = async () => {
    saveGlobalSettings()
    try { await RegisterLinkRouter() }
    catch { 
      showAlertModal(`Failed to unregister:\n\n${err.message || err}`) 
      return
    }
    showSavedNotification("Registered successfully")
}

const unregisterApp = () => {
  showConfirmModal(
    'Unregister LinkRouter',
    'Are you sure you want to unregister LinkRouter from the system?\n\nThis will remove protocol handlers and browser integration.',
    'Unregister',
    'Cancel',
    async () => {
      try {
        await UnregisterLinkRouter();
        // Optional: show success notification
        statusMessage.value = 'Unregistered successfully';
        showSavedNotification(statusMessage.value)
        // setTimeout(() => statusMessage.value = '', 3000);
      } catch (err) {
        // Optional: show error
        showAlertModal(`Failed to unregister:\n\n${err.message || err}`);
        return
      }
      showSettingsModal.value = false
    }
  );
};

// Row selection & context menu
function selectRow(index) {
  selectedRowIndex.value = index;
}

function openContextMenu(event, rule, index) {
  selectRow(index);

  const menuWidth = 180;
  const menuHeight = 250;

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

let clipboardRule = null;

function handleContextAction(action, indexOverride = null) {
  let rule, index;

  if (indexOverride !== null) {
    const item = filteredRules.value[indexOverride];
    if (!item) return;
    rule = item.rule;
    index = item.realIndex;
  } else {
    rule = contextMenu.value.rule;
    index = contextMenu.value.index;
  }

  // Get the actual index in config.value.rules (not filtered list)
  const actualIndex = config.value.rules.findIndex(r =>
    r.regex === rule.regex &&
    r.program === rule.program &&
    r.arguments === rule.arguments
  );

  if (actualIndex === -1) {
    console.warn('Rule not found in config');
    return;
  }

  if (action === 'edit') {
    openEditModal(rule);
  } 
  else if (action === 'delete') {
    const header = `Delete rule #${index + 1}?`
    const message = `${rule.regex}\n↓\n"${basename(rule.program)}" ${rule.arguments}`;
    showConfirmModal(header, message, "Delete", "Cancel", () => {
      config.value.rules.splice(actualIndex, 1);
      if (selectedRowIndex.value === index) {
        selectedRowIndex.value = -1;
      }
      saveConfig();
      saveToUndo();
    });
  }
  else if (action === 'copy') {
    // Deep clone the rule
    clipboardRule = JSON.parse(JSON.stringify(rule));
  }
  else if (action === 'paste') {
    if (clipboardRule) {
      // Insert AFTER the current rule
      config.value.rules.splice(actualIndex + 1, 0, JSON.parse(JSON.stringify(clipboardRule)));
      saveConfig();
      saveToUndo();
    }
  }
  else if (action === 'duplicate') {
    // Clone and insert right after
    const clonedRule = JSON.parse(JSON.stringify(rule));
    config.value.rules.splice(actualIndex + 1, 0, clonedRule);
    saveConfig();
    saveToUndo();
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

async function openConfigInExplorer() {
  if (!configPath.value) return;
  try {
    await window.go.main.App.OpenConfigInExplorer(configPath.value);
  } catch (err) {
    showAlertModal(`Failed to open config in Explorer:\n\n${err.message || err}`);
  }
}

const reportSuccess = ref(false)

let saveNotificationTimeout = null
const notificationKey = ref(0)

const showSavedNotification = (message = "Config saved!") => {
  reportSuccess.value = true
  statusMessage.value = message
  notificationKey.value++

  // Clear previous timer
  if (saveNotificationTimeout) {
    clearTimeout(saveNotificationTimeout)
  }

  // Set new timer from last call
  saveNotificationTimeout = setTimeout(() => {
    reportSuccess.value = false
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

// Alerts
function showAlertModal(message) {
  alertMessage.value = message;
  showAlert.value = true;
}

function showConfirmModal(header, message, okBtn = "OK", cancelBtn = "Cancel", callback) {

  confirmHeader.value = header;
  confirmMessage.value = message;
  confirmOkBtn.value = okBtn;
  confirmCancelBtn.value = cancelBtn;
  onConfirm.value = callback;
  showConfirm.value = true;
}

const handleConfirm = () => {
  if (onConfirm.value) {
    onConfirm.value();
  }
  showConfirm.value = false;
};
</script>
