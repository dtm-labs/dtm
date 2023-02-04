import router from '/@/router'
import { configure, start, done } from 'nprogress'
import { useLayoutStore } from './store/modules/layout'

configure({ showSpinner: false })

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const defaultRoutePath = '/'

router.beforeEach((to) => {
    start()

    const { getMenubar, concatAllowRoutes } = useLayoutStore()

    if (getMenubar.menuList.length === 0) {
        concatAllowRoutes()

        return to.fullPath
    }
})

router.afterEach(() => {
    done()
})
