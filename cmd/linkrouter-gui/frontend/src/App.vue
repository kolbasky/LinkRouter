<template>
  <div class="app">
    <div class="title-bar">
      <div class="title">LinkRouter Config</div>
      <div class="title-bar-buttons">
        <table>
          <tbody>
          <tr>
            <td><button class="close-btn" @click="minimizeWindow">_</button></td>
            <td><button class="close-btn" @click="maximizeWindow">⏹</button></td>
            <td><button class="close-btn" @click="closeWindow">×</button></td>
          </tr>
          </tbody>
        </table>
      </div>
    </div>
    <h1>LinkRouter Config</h1>
    <input v-model="search" placeholder="Search rules..." class="search-input" />
    <div class="container">
      
      
      
      <table class="config-table">
        <thead>
          <tr>
            <th>Pattern (Regex)</th>
            <th>Program</th>
            <!-- <th>Arguments</th> -->
          </tr>
        </thead>
        <tbody>
          <tr v-for="rule in filteredRules" :key="rule.regex">
            <td><code>{{ rule.regex }}</code></td>
            <td><code>{{ basename(rule.program) }}</code></td>
            <!-- <td><code>{{ rule.arguments || '—' }}</code></td> -->
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>


<script setup>
import { WindowMinimise, WindowToggleMaximise, Quit } from '../wailsjs/runtime';
import { ref, computed } from 'vue';
import { GetConfig } from '../wailsjs/go/main/App';

const config = ref({});
const search = ref('');

const closeWindow = () => Quit();
const minimizeWindow = () => WindowMinimise();
const maximizeWindow = () => WindowToggleMaximise();

GetConfig().then(c => config.value = c);


const filteredRules = computed(() => {
  const query = search.value.toLowerCase();
  return config.value.rules?.filter(r => 
    (r.regex && r.regex.toLowerCase().includes(query)) ||
    (r.program && r.program.toLowerCase().includes(query))
  ) || [];
});

function basename(path) {
  if (!path) return '';
  const parts = path.replace(/\\/g, '/').split('/');
  return parts[parts.length - 1] || path;
}
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
