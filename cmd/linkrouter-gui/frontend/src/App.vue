<template>
  <div class="app">
    <!-- Title Bar -->
    <div class="title-bar">
      <div class="title">LinkRouter Config Editor</div>
      <div class="title-bar-buttons">
        <table>
          <tbody>
            <tr>
              <td><button class="close-btn" @click="minimizeWindow" style="scale:0.7">―</button></td> <!-- ― -->
              <td><button class="close-btn" @click="maximizeWindow" style="scale:1.1">▢</button></td> <!--◻ □ ▢-->
              <td><button class="close-btn" @click="closeWindow" style="scale:0.6">╳</button></td> <!-- ╳⨯𐌗𐌢×-->
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Scrollable Content Area -->
    <div class="content">
      <!-- <h1>LinkRouter Config</h1> -->

      <input
        v-model="search"
        placeholder="Search rules..."
        class="search-input"
      />

      <table class="config-table">
        <thead>
          <tr>
            <th>Pattern (Regex)</th>
            <th>Program</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="rule in filteredRules" :key="rule.regex">
            <td><code>{{ rule.regex }}</code></td>
            <td><code>{{ basename(rule.program) }}</code></td>
            <td style="text-align: right; padding: 0.5rem;">
              <button class="edit-btn" @click="openEditModal(rule)">Edit</button>
            </td>
          </tr>
          <tr v-if="filteredRules.length === 0">
            <td colspan="3" style="text-align: center; padding: 2rem; color: #64748b;">
              No rules found
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
      <button class="load-btn" @click="loadConfig">
        Load Config File
      </button>
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
  </div>
</template>


<script setup>
import { WindowMinimise, WindowToggleMaximise, Quit } from '../wailsjs/runtime';
import { ref, computed } from 'vue';
import { GetConfig } from '../wailsjs/go/main/App';
import { OpenFileDialog, LoadConfigFromPath } from '../wailsjs/go/main/App';

const config = ref({});
const search = ref('');

const closeWindow = () => Quit();
const minimizeWindow = () => WindowMinimise();
const maximizeWindow = () => WindowToggleMaximise();
const configPath = ref('');

GetConfig().then(c => config.value = c);

const filteredRules = computed(() => {
  const query = search.value.toLowerCase();
  return config.value.rules?.filter(r => 
    (r.regex && r.regex.toLowerCase().includes(query)) ||
    (r.program && r.program.toLowerCase().includes(query))
  ) || [];
});

const loadConfig = async () => {
  try {
    const filePath = await OpenFileDialog();
    if (!filePath) {
      // User cancelled
      return;
    }

    const newConfig = await LoadConfigFromPath(filePath);
    if (newConfig) {
      config.value = newConfig;
      configPath.value = basename(filePath);
      search.value = ''; // optional: clear search
      // runtime.LogInfo(a.ctx, "Config loaded from: " + filePath); // optional logging
    }
  } catch (err) {
    // You could show a nice error modal here later
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
const originalRule = ref(null); // to find index later

const openEditModal = (rule) => {
  editingRule.value = {
    regex: rule.regex || '',
    program: rule.program || '',
    arguments: rule.arguments || ''
  };
  originalRule.value = rule; // reference to original object
  showEditModal.value = true;
};

const closeEditModal = () => {
  showEditModal.value = false;
  // Reset in case user cancels
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

  // Update the original rule object (Vue reactivity will update the table)
  Object.assign(originalRule.value, editingRule.value);

  closeEditModal();

  // Optional: Later you can send this to Go backend to save to file
  // await SaveConfig(config.value);
};

</script>


<style>
#logo {
  display: block;
  width: 50%;
  height: 50%;
  margin: auto;
  padding: 10% 0 0;
  background-position: center;
  background-repeat: no-repeat;
  background-size: 100% 100%;
  background-origin: content-box;
}
</style>
