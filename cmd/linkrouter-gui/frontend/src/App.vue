<template>
  <div v-if='!launchedInInteractiveMode' class="app" spellcheck="false">
    <!-- Title Bar -->
    <!--WWW <div class="title-bar" @dblclick="maximizeWindow">
      <div class="title">LinkRouter Config Editor</div>
      <div class="title-bar-buttons">
              <button class="titlebar-btn minimize-btn" @click="minimizeWindow">―</button>
              <button class="titlebar-btn maximize-btn" @click="maximizeWindow">◻</button>
              <button class="titlebar-btn close-btn" @click="closeWindow">⨯</button>
    
      </div>
    </div> -->

    <div class="header">
      <div class="search-container">
        <!-- Search Wrapper -->
        <div class="search-wrapper">
          <input
            ref="searchInput"
            v-model="search"
            placeholder="Search rules..."
            class="search-input"
            title="Search rules (Ctrl+F)"
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
        <button 
          class="help-btn-main"
          @click="showHelp"
          title="Show Help (F1)"
        >
          ❓&#65038
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
            <th style="width:5%;">#</th>
            <th style="width:60%">Regex</th>
            <th style="width:30%">Program</th>
            <th style="width:5%;"></th>
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
        <button class="add-rule-btn" @click="openAddRuleModal" title="New rule (Ctrl+N)"><span class="emoji">➕&#65038</span></button>
                <!-- Undo Button -->
        <button 
          class="undo-btn"
          @click="undo()"
          :disabled="!canUndo"
          title="Undo (Ctrl+Z)"
        >
          ↶
        </button>
        <button 
          class="redo-btn"
          @click="redo()"
          :disabled="!canRedo"
          title="Redo (Ctrl+Y)"
        >
          ↷
        </button>
        <button class="reload-btn" @click="reloadConfig(false)" title="Reload config (Ctrl+R)">🔄&#65038</button>
        <button class="load-btn" @click="loadConfig" title="Open config (Ctrl+O)">📂︎&#65038</button>
        <button class="save-btn" @click="saveConfigAs" title="Save config as (Ctrl+S)"><span class="emoji">💾&#65038</span></button>
        <!-- <button class="settings-btn" @click="openSettingsModal" title="Global settings"><span class="emoji">🔧&#65038</span></button> -->
        <button class="settings-btn" @click="openSettingsModal" title="Global settings (Ctrl+G)"><span class="emoji">⛭&#65038</span></button>
      </div>
    </div>

    <!-- Edit Rule Modal -->
    <div v-if="showEditModal" class="modal-overlay"> <!--  @mousedown.self="closeEditModal" -->
      <div class="modal" @click.stop>
        <h2 style="--wails-draggable:drag">Edit Rule
          <sup><button class="help-btn-modal" @click="showHelp" title="Show Help (F1)">❓&#65038</button></sup>
        </h2>
        <div class="modal-form-content">
          <label>Regex Pattern</label>
          <input 
            ref="regexInput"
            v-model="editingRule.regex" 
            class="modal-input" 
            :class="{ 'invalid-regex': regexError }"
            @input="validateRegex"
            placeholder="^https?://(.*\.)?youtube\.com/.*" 
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
              @input='editingRule.program = editingRule.program.replace(/"/g,"")'
            />
            <button class="browse-btn" @click="browseFile('ruleProgram')" title="Browse for program">
              <span class="emoji">📂︎</span>
            </button>
          </div>

          <label>Arguments</label>
          <input
            v-model="editingRule.arguments"
            class="modal-input"
            placeholder="{URL} for URL; $1, $2 etc for captured groups"
          />

          <label
          class="checkbox-label"
          >
            <input 
              type="checkbox" 
              v-model="editingRule.interactive"
              class="interactive-checkbox"
              title="Regex is not used in this case and may be used as a name"
              :disabled="/[$]\d+/.test(editingRule.arguments)"
            />
            <span 
              class="checkbox-hint"
              title="Regex is not used in this case and may be used as a name"
              :class="{ 'disabled': /[$]\d+/.test(editingRule.arguments) }"
            >
              
              {{ /[$]\d+/.test(editingRule.arguments) ? 
                "Add rule as a button to Launcher (disabled becuase of captured groups in arguments)" : 
                "Add rule as a button to Launcher" 
              }}
            </span>
          </label>

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
            <div v-if="testUrl && interactivePrograms.length > 0" class="interactive-buttons">
              <div 
                v-for="(prog, index) in interactivePrograms" 
                :key="index"
                class="interactive-button-row"
              >
                <button
                  class="browser-btn"
                  @click="openTestUrlInBrowser(prog.program, prog.arguments)"
                  :title="`Open in ${prog.name} (Ctrl+${index + 1})`"
                >
                  <span class="browser-btn-icon">
                    <img 
                      :src="getIconUrl(getIcon(prog.program, prog.arguments))" 
                      alt=""
                      class="browser-icon"
                    /></span>
                    &nbsp;{{ prog.name }}
                    <span 
                      v-if="isCtrlPressed" 
                      class="hotkey-hint"
                    >
                      {{ index + 1 }}
                    </span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-buttons">
          <button 
            class="test-rule-btn"
            @click="testRuleLocally"
            title="Test this rule with current test URL (Ctrl+T)"
            :disabled="!testUrl || !editingRule.program || !testResult"
          >
            Test rule
          </button>
          <button class="cancel-btn" @click="closeEditModal" title="Cancel (Esc)">Cancel</button>
          <button class="ok-btn" @click="saveRule" title="Save rule (Ctrl+Enter)">Save</button>
        </div>
      </div>
    </div>

    <!-- Global Settings Modal -->
    <div v-if="showSettingsModal" class="modal-overlay">
      <div class="modal" @click.stop>
        <h2>Global Settings
          <sup><button class="help-btn-modal" @click="showHelp" title="Show Help (F1)">❓&#65038</button></sup>
        </h2>

        <div class="modal-form-content">
          <label>Fallback Browser Path</label>
          <div class="program-input-wrapper">
            <input
              ref="fallbackBrowserInput"
              v-model="editingGlobal.fallbackBrowserPath"
              class="modal-input program-input"
              placeholder="C:\Program Files\Firefox\firefox.exe"
              @input='editingGlobal.fallbackBrowserPath = editingGlobal.fallbackBrowserPath.replace(/"/g,"")'
            />
            <button class="browse-btn" @click="browseFile('fallbackBrowser')" title="Browse for program">
              <span class="emoji">📂︎</span>
            </button>
          </div>

          <label>Fallback Browser Arguments</label>
          <input
            v-model="editingGlobal.fallbackBrowserArgs"
            class="modal-input"
            placeholder="--incognito {URL}"
          />

          <label>
            Interactive Mode
          </label>
          <input type="checkbox" v-model="editingGlobal.interactiveMode" />

          <label>Default Config Editor</label>
          <div class="program-input-wrapper">
            <input
            v-model="editingGlobal.defaultConfigEditor"
            class="modal-input program-input"
            placeholder="notepad.exe"
            @input='editingGlobal.defaultConfigEditor = editingGlobal.defaultConfigEditor.replace(/"/g,"")'
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
            placeholder="logs\linkrouter.log"
            @input='editingGlobal.logPath = editingGlobal.logPath.replace(/"/g,"")'
            />
            <button class="browse-btn" @click="browseFile('logPath')" title="Browse for program">
              <span class="emoji">📂︎</span>
            </button>
          </div>
          
          <label>Supported Protocols (comma-separated)</label>
          <input
          v-model="protocolsInput"
          class="modal-input"
          placeholder="http, https, ssh, mailto"
          @input="protocolsInput = sanitizeProtocols(protocolsInput)"
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
      <button class="context-item" @click="openAddRuleModal">➕︎&nbsp&nbspAdd<br></button>
      <button class="context-item" @click="handleContextAction('edit')">✎&nbsp&nbspEdit<br></button>
      <button class="context-item" @click="handleContextAction('copy')">⧉&nbsp&nbspCopy<br></button>
      <button class="context-item" :disabled="!clipboardRule" @click="handleContextAction('paste')">📋︎&nbsp&nbspPaste<br></button>
      <button class="context-item" @click="handleContextAction('duplicate')">⮻&nbsp&nbspDuplicate<br></button>
      <button class="context-item" @click="handleContextAction('delete')">❌︎&nbsp&nbspDelete<br></button>
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
      <div class="modal-form-content">
        <p>{{ alertMessage }}</p>
      </div>
      <div class="modal-buttons">
        <button class="ok-btn" @click="showAlert = false">OK</button>
      </div>
    </div>
  </div>

  <!-- Confirmation Modal -->
  <div v-if="showConfirm" class="modal-overlay" @click="showConfirm = false">
    <div class="modal confirm-modal" @click.stop>
      <h2>{{ confirmHeader }}</h2>
      <div class="modal-form-content">
        <p>{{ confirmMessage }}</p>
      </div>
      <div class="modal-buttons">
        <button class="cancel-btn" @click="showConfirm = false">{{ confirmCancelBtn }}</button>
        <button class="ok-btn" @click="handleConfirm">{{ confirmOkBtn }}</button>
      </div>
    </div>
  </div>

  <div
    v-if="showHelpOverlay"
    class="help-overlay"
    @click="showHelpOverlay = false"
  >
  <button class="close-help" @click="showHelpOverlay = false">×</button>
    <div class="help-content" @click.stop ref="helpContainer" tabindex="-1">
      <div class="help-body" v-html="helpContent" ref="helpBody" @click="handleHelpLinkClick"></div>
      <div class="help-footer">
        Press F1 • Esc • or click outside to close
      </div>
    </div>
  </div>

<!-- Interactive Mode UI -->
<div v-if="launchedInInteractiveMode" class="interactive-mode">
  <div class="interactive-dialog">
    <div class="url">
      <input v-model="testUrl" class="url-input">
    </input>
    <button class="copy-url-button" @click="copyToClipboard(testUrl, '999', 'url')" title="Copy to clipboard">
      <span class="emoji" v-if="!(copiedIndex === '999' && copiedField === 'url')">📋︎</span>
      <span class="emoji" v-else>✓</span>
    </button>
      
    </div>
    <div class="interactive-buttons-grid">
      <button
        v-for="(prog, index) in interactivePrograms"
        :key="index"
        class="browser-btn"
        @click="openTestUrlInBrowser(prog.program, prog.arguments)"
        :title="`Open in ${prog.name} (Ctrl+${index + 1})`"
      >
        <span class="browser-btn-icon">
          <img 
            :src="getIconUrl(getIcon(prog.program, prog.arguments))" 
            alt=""
            class="browser-icon"
          />
        </span>
        &nbsp;{{ prog.name }}
        <span 
          v-if="isCtrlPressed" 
          class="hotkey-hint"
        >
          {{ index + 1 }}
        </span>
      </button>
    </div>
    
    <button class="create-rule-btn" @click="launchRuleCreation" title="Create new rule for current URL (Ctrl+N)">
      ➕︎ New rule
    </button>
  </div>
</div>

</template>

<script setup>
// import { Fzf } from 'fzf'
import { getIconUrl } from './icons.js';
import { WindowMinimise, WindowToggleMaximise, Quit, WindowUnminimise, BrowserOpenURL, LogInfo } from '../wailsjs/runtime/runtime';
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
  IsValidRegex,
  RegisterLinkRouter,
  UnregisterLinkRouter,
  OpenInFallbackBrowser,
  TestRule,
  ShowCreateRule,
  SpawnNewInstance
} from '../wailsjs/go/main/App';

let interactiveCSSLoaded = false;

// reload config on focus, but with 3sec debounce
let lastFocus = Date.now();
window.addEventListener('focus', () => {
  if (Date.now() - lastFocus > 3000) {
    reloadConfig(true);
  }
  lastFocus = Date.now();
  if (launchedInInteractiveMode.value && launchedInInteractiveModeURL.value) {
    nextTick(() => {
      resizeToDialog('.interactive-dialog');
    })
  }
});

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
    launchedInInteractiveModeURL.value = mode.url
    if (mode.enabled === "true" && mode.url) {
      if (!interactiveCSSLoaded) {
        import('./assets/css/interactive-mode.css');
      interactiveCSSLoaded = true;
      }
      testUrl.value = mode.url
      editingRule.value = {
        regex: guessRegex(mode.url),
        program: '',
        arguments: '"{URL}"',
        interactive: false
      }
      originalRule.value = null
      editingRule.value.regex
      updateTestResult()
      
      nextTick(() => {
        regexInput.value?.focus()
        resizeIntervalId.value = setInterval(() => {
          resizeToDialog('.interactive-dialog');
        }, 5);
      });
      setTimeout(() => {
        clearInterval(resizeIntervalId.value);
      }, 200);
      nextTick(() => {
        centerIntervalId.value = setInterval(() => {
          runtime.WindowCenter();
        }, 5);
      });
      setTimeout(() => {
        clearInterval(centerIntervalId.value);
      }, 200);
    }

    const show = await ShowCreateRule()
    if (show) {
      testUrl.value = mode.url
      editingRule.value = {
        regex: guessRegex(mode.url),
        program: '',
        arguments: '"{URL}"',
        interactive: false
      }
      originalRule.value = null
      showEditModal.value = true
      editingRule.value.regex
      updateTestResult()
      nextTick(() => {
        regexInput.value?.focus()
      })
    }



    // Dirty hack to focus the app
    runtime.WindowMinimise()
    let attempts = 0;
    const maxAttempts = 30;
    const tryUnminimize = () => {
      runtime.WindowUnminimise();
      const currentActive = document.activeElement;
      const isInputFocused = currentActive && 
        currentActive.tagName.toLowerCase() === 'input';
      
      if (isInputFocused) {
        return;
      }
      
      if (++attempts < maxAttempts) {
        setTimeout(tryUnminimize, 50);
      }
    };
    setTimeout(tryUnminimize, 50);
}).catch((err) => {
  showAlertModal(`Loading config failed:\n\n${err.message || err}`)
});

const guessRegex = (url) => {
  try {
    const u = new URL(url)
    return `${u.protocol}//${u.hostname}.*`
  } catch (e) {
    return '.*'
  }
}

const getDialogSize = (d) => {
  const dialog = document.querySelector(d);
  if (!dialog) return { width: 100, height: 100 };
  
  const rect = dialog.getBoundingClientRect();
  // Add some padding for borders/shadow
  return {
    width: Math.ceil(rect.width),
    height: Math.ceil(rect.height)
  };
};

const resizeToDialog = (d) => {
  const { width, height } = getDialogSize(d);
  // Ensure minimum sizes
  const finalWidth = Math.max(width, 100);
  const finalHeight = Math.max(height, 100);
  
  runtime.WindowSetSize(finalWidth, finalHeight);
};  

const launchRuleCreation = async () => {
  runtime.LogInfo(launchedInInteractiveMode.value + " " + launchedInInteractiveModeURL.value)
  await SpawnNewInstance(['--create-rule', '--url', launchedInInteractiveModeURL.value]);
  await runtime.Quit();
};

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
    if (e.shiftKey || e.altKey) {
      return
    }
    const isCtrl  = e.ctrlKey || e.metaKey
    const isPlain = !e.shiftKey && !e.altKey && !e.metaKey && !e.ctrlKey

    // ─────────────────────────────────────────────────────────────
    // Always-working keys
    // ─────────────────────────────────────────────────────────────

    if (e.key === 'Escape') {
      e.preventDefault()

      if (showHelpOverlay.value)            { showHelpOverlay.value = false; restoreFocus(); return }
      if (showAlert.value)                  { showAlert.value = false; restoreFocus(); return }
      if (showConfirm.value)                { showConfirm.value = false; restoreFocus(); return }
      if (showEditModal.value)              { closeEditModal(); restoreFocus(); return }
      if (showSettingsModal.value)          { closeSettingsModal(); restoreFocus(); return }
      if (contextMenu.value.visible)        { closeContextMenu(); restoreFocus(); return }
      if (launchedInInteractiveMode.value)  { Quit() }

      if (document.activeElement !== searchInput.value) {
        searchInput.value?.focus()
      } else {
        rulesContainer.value?.focus()
      }

      return
    }

    if (e.key === 'F1') {
      e.preventDefault();
      if (showHelpOverlay.value) {
        showHelpOverlay.value = false;
        restoreFocus();
      } else {
        rememberFocus();
        showHelp();
      }
      return;
    }

    // ─────────────────────────────────────────────────────────────
    // Priority 2: Enter in modals (confirm / default action)
    // ─────────────────────────────────────────────────────────────

    if (e.key === 'Control' || e.key === 'Meta') {
      isCtrlPressed.value = true;
    }

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
    if (isCtrl && testUrl.value) {
      const keyNum = parseInt(e.key);
      if (!isNaN(keyNum) && keyNum >= 1 && keyNum <= interactivePrograms.value.length + 1) {
        e.preventDefault();
        openTestUrlInBrowser(interactivePrograms.value[keyNum - 1].program, interactivePrograms.value[keyNum - 1].arguments);
        return;
      }
    }

    // Open
    if (isCtrl && e.code === 'KeyO' && !isAnyModalOpen.value) {
      e.preventDefault()
      loadConfig()
      return
    }

    // Global Settings
    if (isCtrl && e.code === 'KeyG' && !isAnyModalOpen.value) {
      e.preventDefault()
      openSettingsModal();
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
        if (launchedInInteractiveMode.value) {
          launchRuleCreation()
          return
        }
        openAddRuleModal()
        return
      }
    }

    if (!isAnyModalOpen.value && (e.key === 'F5' || (isCtrl && e.code === 'KeyR'))) {
      e.preventDefault();
      reloadConfig(false);
      return;
    }

    if (isCtrl && e.code === 'KeyT' && testUrl.value && editingRule.value.program && testResult.value) {
      e.preventDefault();
      testRuleLocally();
      return;
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

setTimeout(() => {
  window.addEventListener('keyup',  (e) => {
    if (e.key === 'Control' || e.key === 'Meta') {
      isCtrlPressed.value = false;
    }
  });
}, 100);

// Reactive state
const canUndo = ref(false);
const canRedo = ref(false);
const config = ref({});
const configPath = ref('');
const copiedIndex = ref(-1)
const copiedField = ref(null)
const isCtrlPressed = ref(false);
const search = ref('');
const rulesContainer = ref(null);
const searchInput = ref(null);
const fallbackBrowserInput = ref(null);
const launchedInInteractiveMode = ref(false);
const launchedInInteractiveModeURL = ref('');
const lastFocusedElement = ref(null);
const resizeIntervalId = ref(null)
const centerIntervalId  = ref(null)

const alertMessage = ref('');
const showAlert = ref(false);
const confirmHeader = ref('');
const confirmMessage = ref('');
const confirmOkBtn = ref('');
const confirmCancelBtn = ref('');
const onConfirm = ref(null);
const showConfirm = ref(false);
const showHelpOverlay = ref(false);
const helpContainer = ref(null);

const showEditModal = ref(false);
const showSettingsModal = ref(false);
const selectedRowIndex = ref(-1);

const editingRule = ref({ regex: '', program: '', arguments: '', interactive: false });
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

const rememberFocus = () => {
  const active = document.activeElement;
  if (!active || 
      active === document.body || 
      active === document.documentElement ||
      active.classList.contains('settings-btn') ||
      active.classList.contains('help-btn-main') ||
      active.classList.contains('help-btn-modal') ||
      active.classList.contains('modal-help-btn')) {
    return;
  }
  lastFocusedElement.value = active;
};

const restoreFocus = () => {
  nextTick(() => {
    if (lastFocusedElement.value) {
      lastFocusedElement.value.focus();
      lastFocusedElement.value = null;
      return
    }
    if (showHelpOverlay.value) {
      helpContainer.value?.focus();
    } else if (showEditModal.value) {
      regexInput.value?.focus();
    } else if (showSettingsModal.value) {
      fallbackBrowserInput.value?.focus();
    } else {

      searchInput.value?.focus();
    }
  });
};

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
  editingRule.value = { regex: '.*', program: '', arguments: '"{URL}"', interactive: false };
  originalRule.value = null;
  showEditModal.value = true;
  closeContextMenu();
  nextTick(() => {
    regexInput.value?.focus()
  });
};

const openEditModal = (rule) => {
  rememberFocus();
  editingRule.value = {
    regex: rule.regex || '',
    program: rule.program || '',
    arguments: rule.arguments || '',
    interactive: rule.interactive || false
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
    editingRule.value = { regex: '', program: '', arguments: '', interactive: false };
    originalRule.value = null;
    regexError.value = null;
  }, 300);
  rulesContainer.value?.focus()
  if (launchedInInteractiveMode.value) {
    runtime.Quit()
  }
};

const interactivePrograms = computed(() => {
  const programs = [];
  if (config.value.global?.fallbackBrowserPath) {
    programs.push({
      program: config.value.global.fallbackBrowserPath,
      arguments: config.value.global.fallbackBrowserArgs || '{URL}',
      name: 'default'
    });
  }
  programs.push(...config.value.rules
    .filter(rule => rule.interactive)
    .map(rule => ({
      program: rule.program,
      arguments: rule.arguments,
      name: getProgramName(rule.program)
    }))
  );
  return programs;
});

// Helper function to extract program name from path
const getProgramName = (programPath) => {
  if (!programPath) return 'Unknown';
  
  // Remove quotes if present
  let cleanPath = programPath;
  if (cleanPath.startsWith('"') && cleanPath.endsWith('"')) {
    cleanPath = cleanPath.slice(1, -1);
  }
  
  // Get filename without extension
  const filename = cleanPath.split(/[\\/]/).pop() || cleanPath;
  return filename.replace(/\.[^/.]+$/, ''); // Remove extension
};

const getIcon = (programPath, args) => {
  const pathLower = (programPath || '').toLowerCase();
  const argsLower = (args || '').toLowerCase();
  
  if (argsLower.includes('--incognito') || argsLower.includes('-private-window')) {
    return 'incognito';
  }
  
  if (pathLower.includes('chrome')) return 'chrome';
  if (pathLower.includes('firefox')) return 'firefox';
  if (pathLower.includes('brave')) return 'brave';
  if (pathLower.includes('edge')) return 'edge';
  if (pathLower.includes('opera')) return 'opera';
  if (pathLower.includes('vivaldi')) return 'vivaldi';
  if (pathLower.includes('yandex')) return 'yandex';
  
  return 'generic';
};

const openTestUrlInBrowser = async (path = config.value.global.fallbackBrowserPath, args = "\"{URL}\"") => {
  if (!testUrl.value?.trim()) return;
  try {
    await OpenInFallbackBrowser(path, args, '"' + testUrl.value.trim() + '"');
  } catch (err) {
    runtime.LogError("Failed to open URL:", err);
  }
  if (launchedInInteractiveMode.value) {
    runtime.Quit()
  }
};

const openExternal = (url) => {
  showAlert(url)
  BrowserOpenURL(url)
}

const testRuleLocally = async () => {
  if (!testUrl.value || !editingRule.value.program) {
    return;
  }

  try {
    const unquotedProgram = editingRule.value.program.replace(/^"+|^'+|"+$|'+$/g, "");
    await TestRule(
      { ...editingRule.value, program: unquotedProgram },
      testUrl.value
    )
  } catch(err) {
    showAlertModal(`Failed to launch:\n\n${err.message || err}`);
  }
}

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
  rememberFocus();
  editingGlobal.value = {
    fallbackBrowserPath: config.value.global?.fallbackBrowserPath || '',
    fallbackBrowserArgs: config.value.global?.fallbackBrowserArgs || '',
    defaultConfigEditor: config.value.global?.defaultConfigEditor || '',
    logPath: config.value.global?.logPath || '',
    interactiveMode: config.value.global?.interactiveMode || false,
    supportedProtocols: [...(config.value.global?.supportedProtocols || [])]
  };
  originalGlobal.value = config.value.global;
  protocolsInput.value = editingGlobal.value.supportedProtocols.join(',');
  showSettingsModal.value = true;
  nextTick(() => {
    fallbackBrowserInput.value?.focus()
  });
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
    const cleanProtocols = protocolsInput.value
      .replace(/\s+/g, ',')
      .replace(/,+/g, ',')
      .replace(/[^a-zA-Z0-9+.,-]/g, '')
      .replace(/^,|,$/g, ''); 
    editingGlobal.value.supportedProtocols = cleanProtocols
      .split(',')
      .map(s => s.trim())
      .filter(s => s && /^[a-z][a-z0-9+.-]*$/i.test(s));

    if (!config.value.global) {
      config.value.global = {};
    }

    Object.assign(config.value.global, editingGlobal.value);
    try { await RegisterLinkRouter() }
    catch { 
      showAlertModal(`Failed to unregister:\n\n${err.message || err}`) 
      return
    }
    closeSettingsModal();
  } catch (err) {
    showAlertModal(`Failed to save settings:\n\n${err.message || err}`);
  }
    saveConfig();
    saveToUndo();
};

const sanitizeProtocols = (p) => {
  p = p.replace(/[^a-zA-Z0-9+.,-]/g, '')
  p = p.replace(/,+/g,',')
  p = p.replace(/(^|,)[^a-zA-Z]/g,'$1')
  return p
}

const registerApp = async () => {
    saveGlobalSettings()
    try { await RegisterLinkRouter(true) }
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
  
  canUndo.value = historyIndex.value > 0;
  canRedo.value = false;
}

const undo = () => {
  if (historyIndex.value <= 0) return

  historyIndex.value--
  config.value = JSON.parse(JSON.stringify(history.value[historyIndex.value]))
  saveConfig();
  canRedo.value = true;
  canUndo.value = historyIndex.value > 0;
}

const redo = () => {
  if (historyIndex.value >= history.value.length - 1) return

  historyIndex.value++
  config.value = JSON.parse(JSON.stringify(history.value[historyIndex.value]))
  saveConfig();
  canUndo.value = true;
  canRedo.value = historyIndex.value < history.value.length - 1;
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


const handleHelpLinkClick = (event) => {
  let target = event.target
  while (target && target.tagName !== 'A') {
    target = target.parentElement
  }

  if (target && target.tagName === 'A' && target.href) {
    event.preventDefault()
    event.stopPropagation()

    const url = target.getAttribute('href')
    if (url && url !== '#' && !url.startsWith('javascript:')) {
      BrowserOpenURL(url)
    } else if (target.hasAttribute('data-external')) {
      const customUrl = target.getAttribute('data-external')
      if (customUrl) BrowserOpenURL(customUrl)
    }
  }
}

/* help */
const showHelp = () => {
  rememberFocus();
  showHelpOverlay.value = true;
  
  nextTick(() => {
    if (helpContainer.value) {
      helpContainer.value?.focus();
      helpContainer.value.scrollTop = 0;
    }
  });
};

const helpContent = computed(() => {
  if (showEditModal.value) {
    return `
      <h2>Working with Rules</h2>
      <p><strong>Regex Pattern</strong><br>
         If the incoming URL matches this regular expression, the specified program will be launched and further rule processing will stop.
      </p>
      <p>Examples:</p>
      <ul>
        <li><code>https://store.steampowered.com.*</code> — simple pattern</li>
        <li><code>https?://(.*\\.)?youtube\\.com</code> — more flexible</li>
        <li><code>mailto:(.*@(company1|company2)\\.(com|ru))</code> — using capture groups</li>
      </ul>

      <p><strong>Program</strong><br>
         Full path to the application to launch when the regex matches.<br>
         If you provide only a filename, LinkRouter will look for it in <code>%PATH%</code>.
      </p>

      <p><strong>Arguments</strong><br>
         Command-line arguments passed to the program.<br>
         <code>{URL}</code> is replaced with the full incoming URL.<br>
         <code>$1</code>, <code>$2</code>, … are replaced with contents of capture groups.<br>
         Recommendation: always wrap the URL in double quotes, e.g. <code>"{URL}"</code> or <code>"mailto:$1"</code>.
      </p>

      <p><strong>Add rule as a button to Launcher</strong><br>
        When enabled, this rule appears as a button in the Launcher and the rule-creation dialog when Test URL is not empty.<br>
        <br>
        Use it to turn LinkRouter into a quick app/browser selector:<br>
        • No matching rule → dialog opens with pre-filled URL<br>
        • Click button → opens link in the chosen program<br>
        • Keyboard: <code>Ctrl+1</code>, <code>Ctrl+2</code>, …<br>
        <br>
        Note: Since no regex matching happens, captured groups are not expanded in <code>Arguments</code> - only <code>{URL}</code> is replaced. Place these rules in the bottom of rule list or fill regex with something unique e.g "Interactive rule #1" to not interfere with normal rule flow.
      </p>

      <p><strong>Test URL</strong><br>
         Type or paste any link here to see a live test against your current regex.<br>
         Green border = match, red border = no match. Very useful while writing new rules.
      </p>

      <p><strong>Test rule button</strong><br>
      <code>Test rule</code> launches the specified program with the current arguments and test URL.
      </p>

      <hr>
      <h2>Keyboard shortcuts in this modal</h2>
      <ul>
        <li><code>Ctrl + Enter</code> — Save rule and close modal (config is saved automatically)</li>
        <li><code>Ctrl + T</code> — Test rule</li>
        <li><code>Ctrl + O</code> — Open test URL in fallback browser</li>
        <li><code>Esc</code> — Close modal</li>
      </ul>
    `
  }

  if (showSettingsModal.value) {
    return `
      <h2>Global Settings</h2>

      <p><strong>Fallback Browser</strong><br>
         If no rule matches the incoming URL, it will be opened using this program.
      </p>

      <p><strong>Fallback Arguments</strong><br>
         Arguments passed to the fallback browser.<br>
         Recommendation: wrap the URL in quotes, e.g. <code>--private-window "{URL}"</code>
      </p>

      <p><strong>Interactive Mode</strong><br>
         When enabled and no rule matches, instead of opening URL in fallback browser:<br>
         • the rule creation dialog opens automatically<br>
         • regex and test URL fields are pre-filled<br>
         • you can press <code>Open in browser</code> or <code>Ctrl+O</code> to skip rule creation and use the fallback browser
      </p>

      <p><strong>Default Config Editor</strong><br>
         Program used to open the config file when <code>linkrouter.exe</code> is double-clicked.<br>
         Usually not needed — you're already using this graphical config editor.
      </p>

      <p><strong>Log Path</strong><br>
         Path to the log file (relative or absolute).<br>
         LinkRouter logs all URL handling actions here.<br>
         Please attach this file when reporting issues on GitHub.
      </p>

      <p><strong>Supported Protocols</strong><br>
         Comma-separated list of protocols LinkRouter should register/handle.<br>
         Example: <code>http, https, ssh, mailto, magnet</code>
      </p>

      <hr>
      <h2>Register / Unregister</h2>
      <p>
        These buttons register or unregister LinkRouter as the default handler for the listed protocols in the operating system.<br>
        The <strong>Register</strong> button also opens Windows settings so you can confirm/select LinkRouter as the default app for those protocols.<br>
        <strong>In most cases you don't need to use these buttons manually</strong> — pressing <strong>Save</strong> usually handles everything automatically.
      </p>
    `
  }

  // Default: main screen
  return `
    <h2>LinkRouter Config Editor</h2>
    <p>
      <a href="https://github.com/kolbasky/LinkRouter">GitHub repository</a>
    </p>
    <p>
      LinkRouter is a lightweight Windows application that routes URLs/links to specific programs according to regex-based rules.<br>
      This GUI editor helps you create, edit, search and validate rules, test regex matching in real time, and supports an interactive mode — when no rule matches, it automatically opens a dialog with suggested regex and the clicked URL pre-filled.
    </p>
    <p>
      • Double-click a rule (or press Enter) to edit it<br>
      • Right-click for additional options (copy, paste, duplicate, delete…)<br>
      • Rules are evaluated from top to bottom — processing stops at the first match
    </p>
    <p>
      The search field looks inside <code>regex</code>, <code>program</code> and <code>arguments</code> fields.<br>
      It also tests the search string against each regex — so you can quickly find which rule would match a particular URL.
    </p>
    <p>
      All changes are saved automatically.<br>
      Use the <strong>undo</strong>/<strong>redo</strong> buttons or <code>Ctrl+Z</code> / <code>Ctrl+Y</code> to revert actions.
    </p>
    <p>
      The configuration is stored in JSON format and can also be edited manually in any text editor.<br>
      The current config path is shown at the bottom of the window — click it to open the file in Explorer.
    </p>

    <hr>
    <h2>Buttons</h2>
    <ul>
      <li><code>➕︎&#65038</code> — Create new rule (<code>Ctrl+N</code>)</li>
      <li><code>↶︎&#65038</code> — Undo last change (<code>Ctrl+Z</code>)</li>
      <li><code>↷︎&#65038</code> — Redo (<code>Ctrl+Y</code>)</li>
      <li><code>🔄︎&#65038</code> — Reload config from disk (<code>Ctrl+R</code> or <code>F5</code>)</li>
      <li><code>📂︎&#65038</code> — Open / load different config file (<code>Ctrl+O</code>)</li>
      <li><code>💾︎&#65038</code> — Save config as… (<code>Ctrl+S</code>)</li>
      <li><code>⛭︎&#65038</code> — Global settings (<code>Ctrl+G</code>)</li>
    </ul>

    <h2>Keyboard shortcuts – main window</h2>
    <ul>
      <li><code>Ctrl+N</code> — create new rule</li>
      <li><code>Ctrl+C</code> / <code>Ctrl+V</code> — copy / paste selected rule</li>
      <li><code>Ctrl+D</code> — duplicate selected rule</li>
      <li><code>Ctrl+Z</code> / <code>Ctrl+Y</code> — undo / redo</li>
      <li><code>Ctrl+F</code>, <code>Ctrl+L</code>, <code>/</code> — focus search field</li>
      <li><code>Ctrl+S</code> — save current config</li>
      <li><code>Ctrl+R</code> / <code>F5</code> — reload config</li>
      <li><code>Ctrl+O</code> — open "Load config" dialog</li>
      <li><code>Ctrl+G</code> — open Global Settings</li>
      <li><code>↑</code> / <code>↓</code> — navigate between rules</li>
      <li><code>Enter</code> — edit selected rule</li>
      <li><code>Delete</code> / <code>Backspace</code> — delete selected rule</li>
    </ul>

    <p>
      Each modal dialog has its own additional shortcuts — see the help text inside that dialog.
    </p>
  `
})

</script>
