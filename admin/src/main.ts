import { createApp } from 'vue'
import App from './App.vue'
import router from '/@/router/index'
import { pinia } from '/@/store'
import '/@/permission'

import '/@/assets/css/index.css'
import 'virtual:svg-icons-register'

const app = createApp(App)
app.use(router)
app.use(pinia)
app.mount('#app')
