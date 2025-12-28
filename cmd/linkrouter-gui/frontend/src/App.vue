<template>
  <div class="app">
    <!-- Title Bar -->
    <div class="title-bar">
      <div class="title">LinkRouter Config Editor</div>
      <div class="title-bar-buttons">
        <table>
          <tbody>
            <tr>
              <td><button class="titlebar-btn minimize-btn" @click="minimizeWindow">―</button></td> <!-- ― -->
              <td><button class="titlebar-btn maximize-btn" @click="maximizeWindow">◻</button></td> <!--◻ □ ▢-->
              <td><button class="titlebar-btn close-btn" @click="closeWindow">⨯</button></td> <!-- ╳⨯𐌗𐌢×-->
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

    <!-- <table class="config-table-header">
      <thead>
        <tr>
          <th style="width: 65%;">Regex</th>
          <th style="width: 35%;">Program</th>
        </tr>
      </thead>
    </table> -->

    <!-- Scrollable Content Area -->
    <div class="content">
      <!-- Table wrapper for scrolling + sticky header -->
        <table class="config-table">
          <thead>
            <tr>
              <th style="width:3%; text-align: center;">#</th>
              <th style="width:64%">Regex</th>
              <th style="width:33%">Program</th>
            </tr>
          </thead>
          <tbody>
            <tr 
              v-for="item in filteredRules" 
              :key="item.originalIndex"
              class="rule-row"
              @dblclick="openEditModal(item.rule)"
            >
              <td class="index-cell">{{ item.originalIndex }}</td>
              <td>
                <div class="code-wrapper">
                  <code class="rule-row-tag">{{ item.rule.regex }}</code>
                  <button class="copy-btn" @click.stop="copyToClipboard(item.rule.regex)" title="Copy to clipboard">
                    📋
                  </button>
                </div>
              </td>
              <td>
                <div class="code-wrapper">
                  <code class="rule-row-tag">{{ basename(item.rule.program) }}</code>
                  <button class="copy-btn" @click.stop="copyToClipboard(item.rule.program)" title="Copy to clipboard">
                    📋
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="filteredRules.length === 0">
              <td colspan="2" style="text-align: center; padding: 2rem; color: #64748b;">
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
        <span v-if="configPath" class="config-path">• {{ configPath }}</span>
      </div>

      <div class="button-group">
        <button class="save-btn" @click="saveConfigAs">
          💾
        </button>
        <button class="settings-btn" @click="openSettingsModal" title="Global Settings">
          ⚙️
        </button>
        <button class="load-btn" @click="loadConfig">
          Load Config File
        </button>
      </div>
    </div>

    <div v-if="showEditModal" class="modal-overlay" @click="closeEditModal">
      <div class="modal" @click.stop>
        <h2>Edit Rule</h2>

        <label>Pattern (Regex)</label>
        <input v-model="editingRule.regex" class="modal-input" placeholder="e.g. ^https?://(.*\\.)?youtube\\.com/.*" />

        <label>Program</label>
        <input v-model="editingRule.program" class="modal-input" placeholder="C:\\Program Files\\App\\app.exe" />

        <label>Arguments (optional)</label>
        <input v-model="editingRule.arguments" class="modal-input" placeholder="--url {url}" />

        <div class="modal-buttons">
          <button class="cancel-btn" @click="closeEditModal">Cancel</button>
          <button class="ok-btn" @click="saveRule">OK</button>
        </div>
      </div>
    </div>

    <div v-if="showSettingsModal" class="modal-overlay" @click="closeSettingsModal">
      <div class="modal" @click.stop>
        <h2>Global Settings</h2>

        <label>Fallback Browser Path</label>
        <input v-model="editingGlobal.fallbackBrowserPath" class="modal-input" placeholder="e.g. C:\\Program Files\\Firefox\\firefox.exe" />

        <label>Fallback Browser Arguments</label>
        <input v-model="editingGlobal.fallbackBrowserArgs" class="modal-input" placeholder="e.g. -private-window {url}" />

        <label>Default Config Editor</label>
        <input v-model="editingGlobal.defaultConfigEditor" class="modal-input" placeholder="e.g. notepad.exe" />

        <label>Log Path</label>
        <input v-model="editingGlobal.logPath" class="modal-input" placeholder="e.g. C:\\logs\\linkrouter.log" />

        <label>Supported Protocols (comma-separated)</label>
        <input v-model="protocolsInput" class="modal-input" placeholder="e.g. http,https,ftp" />

        <div class="modal-buttons">
          <button class="cancel-btn" @click="closeSettingsModal">Cancel</button>
          <button class="ok-btn" @click="okGlobalSettings">Ok</button>
        </div>
      </div>
    </div>

  </div>
</template>


<script setup>
import { WindowMinimise, WindowToggleMaximise, Quit, LogInfo } from '../wailsjs/runtime';
import { ref, computed } from 'vue';
import { OpenFileDialog, LoadConfigFromPath, SaveConfig, SaveConfigAs, GetConfig, GetCurrentConfigPath } from '../wailsjs/go/main/App';

async function loadInitialData() {
  try {
    const [cfg, path] = await Promise.all([
      GetConfig(),
      GetCurrentConfigPath()
    ]);
    config.value = cfg;
    configPath.value = path;
  } catch (err) {
    console.error("Failed to load initial config/path:", err);
    config.value = { rules: [], global: {} };
    configPath.value = '';
  }
}

loadInitialData();

const config = ref({});
const search = ref('');
const closeWindow = () => Quit();
const minimizeWindow = () => WindowMinimise();
const maximizeWindow = () => WindowToggleMaximise();
const configPath = ref('');
const showSettingsModal = ref(false);
const protocolsInput = ref('');
const editingGlobal = ref({
  fallbackBrowserPath: '',
  fallbackBrowserArgs: '',
  defaultConfigEditor: '',
  logPath: '',
  supportedProtocols: []
});
const originalGlobal = ref(null);

// const protocolsInput = computed({
//   get() {
//     return config.value.global?.supportedProtocols?.join(', ') || '';
//   },
//   set(val) {
//     config.value.global.supportedProtocols = val
//       .split(',')
//       .map(s => s.trim())
//       .filter(s => s.length > 0);
//   }
// });

const openSettingsModal = () => {
  editingGlobal.value = {
    fallbackBrowserPath: config.value.global?.fallbackBrowserPath || '',
    fallbackBrowserArgs: config.value.global?.fallbackBrowserArgs || '',
    defaultConfigEditor: config.value.global?.defaultConfigEditor || '',
    logPath: config.value.global?.logPath || '',
    supportedProtocols: [...(config.value.global?.supportedProtocols || [])]
  };
  originalGlobal.value = config.value.global;
  protocolsInput.value = editingGlobal.value.supportedProtocols.join(', ');
  showSettingsModal.value = true;
};

const closeSettingsModal = () => {
  showSettingsModal.value = false;
  setTimeout(() => {
    editingGlobal.value = { fallbackBrowserPath: '', fallbackBrowserArgs: '', defaultConfigEditor: '', logPath: '', supportedProtocols: [] };
    originalGlobal.value = null;
  }, 300);
};

const okGlobalSettings = async () => {
  try {
    // Update editingGlobal with the input value
    editingGlobal.value.supportedProtocols = protocolsInput.value
      .split(',')
      .map(s => s.trim())
      .filter(s => s.length > 0);
    
    // Now copy to config
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
};


Promise.all([
  GetConfig(),
  GetCurrentConfigPath()
]).then(([cfg, path]) => {
  // Add IDs to rules if they don't have them
  cfg.rules = (cfg.rules || []).map((rule, index) => ({
    ...rule,
    id: rule.id || `rule-${index}-${Date.now()}`
  }));
  config.value = cfg;
  configPath.value = path;
});

const filteredRules = computed(() => {
  const query = search.value.toLowerCase();
  const rules = config.value.rules || [];
  
  if (query === '') {
    return rules.map((rule, index) => ({
      rule,
      originalIndex: index + 1, // 1-based
      displayIndex: index + 1
    }));
  }
  
  const filtered = rules
    .map((rule, originalIndex) => ({
      rule,
      originalIndex: originalIndex + 1,
      matchesRegex: rule.regex?.toLowerCase().includes(query) || false,
      matchesProgram: rule.program?.toLowerCase().includes(query) || false
    }))
    .filter(item => item.matchesRegex || item.matchesProgram);
    
  // Return with display index (1-based in filtered list)
  return filtered.map((item, displayIndex) => ({
    ...item,
    displayIndex: displayIndex + 1
  }));
});

const loadConfig = async () => {
  try {
    const filePath = await OpenFileDialog();
    if (!filePath) {
      return;
    }

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

function basename(path) {
  if (!path) return '';
  const parts = path.replace(/\\/g, '/').split('/');
  return parts[parts.length - 1] || path;
}


const showEditModal = ref(false);
const editingRule = ref({
  regex: '',
  program: '',
  arguments: ''
});
const originalRule = ref(null);

const openEditModal = (rule) => {
  editingRule.value = {
    regex: rule.regex || '',
    program: rule.program || '',
    arguments: rule.arguments || ''
  };
  originalRule.value = rule;
  showEditModal.value = true;
};

const closeEditModal = () => {
  showEditModal.value = false;
  setTimeout(() => {
    editingRule.value = { regex: '', program: '', arguments: '' };
    originalRule.value = null;
  }, 300);
};

const saveRule = () => {
  if (!editingRule.value.regex || !editingRule.value.program) {
    alert('Regex and Program are required!');
    return;
  }

  Object.assign(originalRule.value, editingRule.value);

  closeEditModal();
};

const saveConfigAs = async () => {
  try {
    const newPath = await SaveConfigAs(config.value);
    if (newPath) {
      configPath.value = newPath;
    }
  } catch (err) {
    alert('Failed to save config as: ' + err);
    console.error(err);
  }
};

async function copyToClipboard(text) {
  if (!text) return;
  try {
    await navigator.clipboard.writeText(text);
    // Optional: show a tiny toast or change button icon briefly
  } catch (err) {
    console.error('Failed to copy:', err);
    // Fallback for older browsers (not needed in Wails, but safe)
    const textarea = document.createElement('textarea');
    textarea.value = text;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    document.body.removeChild(textarea);
  }
}

</script>
