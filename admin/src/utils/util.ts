import { useRoute } from 'vue-router'
import { IMenubarList } from '../type/store/layout'

export const findCurrentMenubar = (menuList: IMenubarList[], root?: boolean) => {
    const route = useRoute()
    let arr: IMenubarList[] | IMenubarList = []
    for (let i = 0; i < menuList.length; i++) {
        const v = menuList[i]
        const usePath = v.meta.activeMenu || v.redirect || v.path
        const pos = usePath.lastIndexOf('/')
        const rootPath = pos == 0 ? usePath : usePath.substring(0, pos)
        if (route.path.indexOf(rootPath) !== -1) {
            if (!root) {
                arr = v.children as IMenubarList[]
            } else {
                arr = v
            }
            break
        }
    }

    return arr
}

export const sleep = async(ms: number) => {
    return new Promise(resolve => setTimeout(resolve, ms))
}