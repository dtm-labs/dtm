<template>
    <a-layout>
        <a-layout-sider width="200" style="background: #fff">
            <Sidebar />
        </a-layout-sider>
        <a-layout style="padding: 0 24px 24px">
            <div v-if="layout.dtmVersion && layout.dtmVersion != dashVer" style="color:#f00"> !!! admin version: {{ dashVer }} != dtm version: {{ layout.dtmVersion }}. </div>
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
import { findCurrentMenubar } from '../utils/util'
import { computed, onMounted } from 'vue'


const dashVer = import.meta.env.VITE_ADMIN_VERSION
const route = useRoute()
const layout = useLayoutStore()

const mainNav = computed(() => {
    const currentMenubar = findCurrentMenubar(layout.getMenubar.menuList, true)
    return currentMenubar?.meta.title
})

const subNav = computed(() => {
    let subNav = ''
    const currentMenubar = findCurrentMenubar(layout.getMenubar.menuList, true)
    currentMenubar.children?.forEach(v => {
        if (route.path.indexOf(v.path) !== -1) {
            subNav = v.meta.title
        }
    })

    return subNav
})

const page = computed(() => {
    let page = ''
    const currentMenubar = findCurrentMenubar(layout.getMenubar.menuList, true)
    currentMenubar.children?.forEach(v => {
        v.children?.forEach(vv => {
            if (route.path == vv.path) {
                page = vv.meta.title
            }
        })
    })

    return page
})

onMounted(() => {
    layout.loadDtmVersion()
})

</script>

<style lang="postcss" scoped>
.ant-layout.ant-layout-has-sider {
    min-height: calc(100vh - 64px);
}
</style>
