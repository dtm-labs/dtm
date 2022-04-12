import { defineStore } from 'pinia';
import { allowRouter } from '/@/router';
import { ILayout, IMenubar, IMenubarList, IStatus } from '/@/type/store/layout';
import { findCurrentMenubar } from '/@/utils/util';

export const useLayoutStore = defineStore({
    id: 'layout',
    state: ():ILayout => ({
        menubar: {
            menuList: []
        },
        status: {
            isLoading: false
        },
    }),
    getters: {
        getMenubar(): IMenubar {
            return this.menubar
        },
        getStatus(): IStatus {
            return this.status
        },
    },
    actions: {
        setRoutes(data: Array<IMenubarList>): void {
            this.menubar.menuList = data
        },
        concatAllowRoutes(): void {
            allowRouter.reverse().forEach(v => this.menubar.menuList.unshift(v))
        },
    }
})
