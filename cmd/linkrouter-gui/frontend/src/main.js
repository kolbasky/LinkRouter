import {createApp} from 'vue'
import App from './App.vue'
import './style.css';

// Disable the browser's native context menu — we render out own
window.addEventListener('contextmenu', (e) => e.preventDefault());

createApp(App).mount('#app')
