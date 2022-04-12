<template>
    <a-layout style="height: 730px">
        <a-layout-sider width="200" style="background: #fff">
            <Sidebar />
        </a-layout-sider>
        <a-layout style="padding: 0 24px 24px">
            <a-breadcrumb style="margin: 16px 0">
                <a-breadcrumb-item>{{ mainNav }}</a-breadcrumb-item>
                <a-breadcrumb-item>{{ subNav }}</a-breadcrumb-item>
                <a-breadcrumb-item>{{ page }}</a-breadcrumb-item>
            </a-breadcrumb>
            <a-layout-content
            :style="{ background: '#fff', padding: '24px', margin: 0, minHeight: '280px' }"
            >
                <Content />
            </a-layout-content>
        </a-layout>
    </a-layout>
</template>

<script setup lang='ts'>
import Sidebar from './components/sidebar.vue'
import Content from './components/content.vue'
import { useRoute } from 'vue-router'
import { useLayoutStore } from '../store/modules/layout'
import { IMenubarList } from '../type/store/layout'
import { findCurrentMenubar } from '../utils/util'
import { computed, onMounted, ref } from 'vue'

const route = useRoute()
const { getMenubar } = useLayoutStore()

const mainNav = computed(() => {
    const currentMenubar = findCurrentMenubar(getMenubar.menuList, true)
    return currentMenubar?.meta.title
})

const subNav = computed(() => {
    let subNav = ''
    const currentMenubar = findCurrentMenubar(getMenubar.menuList, true)
    currentMenubar.children?.forEach(v => {
        if (route.path.indexOf(v.path) !== -1) {
            subNav = v.meta.title
        }
    })

    return subNav
})

const page = computed(() => {
    let page = ''
    const currentMenubar = findCurrentMenubar(getMenubar.menuList, true)
    currentMenubar.children?.forEach(v => {
        v.children?.forEach(vv => {
            if (route.path == vv.path) {
                page = vv.meta.title
            }
        })
    })   

    return page
})
</script>
