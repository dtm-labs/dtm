import { defineStore } from 'pinia'
import { allowRouter } from '/@/router'
import { ILayout, IMenubar, IMenubarList, IStatus } from '/@/type/store/layout'
import { getDtmVersion } from '/@/api/api_dtm'

export const useLayoutStore = defineStore({
    id: 'layout',
    state: (): ILayout => ({
        menubar: {
            menuList: []
        },
        status: {
            isLoading: false
        },
        dtmVersion: '',
        globalError: ''
    }),
    getters: {
        getMenubar(): IMenubar {
            return this.menubar
        },
        getStatus(): IStatus {
            return this.status
        }
    },
    actions: {
        setRoutes(data: Array<IMenubarList>): void {
            this.menubar.menuList = data
        },
        setGlobalError(err: string) {
            this.globalError = err
        },
        concatAllowRoutes(): void {
            allowRouter.reverse().forEach(v => this.menubar.menuList.unshift(v))
        },
        async loadDtmVersion(): Promise<void> {
            const { data: { version } } = await getDtmVersion()
            this.dtmVersion = version
            console.log('dtm version: ', this.dtmVersion)
        }
    }
})
