import { createApp } from 'vue'
import App from './App.vue'
import router from '/@/router/index'
import { pinia } from '/@/store'
import '/@/permission'

import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/antd.css'
import 'virtual:svg-icons-register'
import SvgIcon from '/@/components/SvgIcon/index.vue'

const app = createApp(App)
app.use(Antd)
app.use(router)
app.use(pinia)
app.component('SvgIcon', SvgIcon)
app.mount('#app')
