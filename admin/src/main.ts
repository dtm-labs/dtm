import { createApp } from 'vue'
import App from './App.vue'
import router from '/@/router/index'
import { pinia } from '/@/store'
import { useLayoutStore } from '/@/store/modules/layout'
import '/@/permission'

import 'ant-design-vue/dist/antd.css'
import '/@/assets/css/index.css'
import 'virtual:svg-icons-register'

const app = createApp(App)
app.use(router)
app.use(pinia)
app.mount('#app')

window.onunhandledrejection = (ev: PromiseRejectionEvent) => {
    showAlert(ev.reason.stack || ev.reason.message)
}
window.onerror = err => {
    if (typeof err === 'string') {
        return showAlert(err)
    }
    showAlert(JSON.stringify(err))
}

function showAlert(msg: string) {
    const layout = useLayoutStore()
    if (!layout.globalError) {
        layout.setGlobalError(msg)
    }
}
