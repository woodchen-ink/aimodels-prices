import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'
import axios from 'axios'

// 配置 axios
axios.defaults.withCredentials = true

const app = createApp(App)

app.use(ElementPlus)
app.use(router)
app.mount('#app') 