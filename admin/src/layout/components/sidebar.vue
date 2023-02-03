<template>
    <a-menu
        v-model:selectedKeys="activeMenu"
        v-model:openKeys="openKeys"
        mode="inline"
        :style="{ height: '100%', borderRight: 0 }"
        @select="onOpenChange"
    >
        <a-sub-menu v-for="v in filterSubMenubarData" :key="v.path">
            <template #title>
                <span>
                    {{ v.meta.title }}
                </span>
            </template>
            <a-menu-item v-for="vv in v.children" :key="vv.path">{{ vv.meta.title }}</a-menu-item>
        </a-sub-menu>
    </a-menu>
</template>

<script setup lang='ts'>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useLayoutStore } from '/@/store/modules/layout'
import { IMenubarList } from '/@/type/store/layout'
import { findCurrentMenubar } from '/@/utils/util'

const route = useRoute()
const router = useRouter()
const { getMenubar } = useLayoutStore()

const filterSubMenubarData = computed(() => {
    return findCurrentMenubar(getMenubar.menuList) as IMenubarList[]
}) 

const activeMenu = computed({
    get: () => {
        return [route.path]
    },
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    set: (val) => {
        // do nothing, just for eliminate warn
    }
})

const openKeys = computed({
    get: () => {
        const pos = route.path.lastIndexOf('/')
        return [route.path.substring(0, pos)]
    }, 
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    set: (val) => {
        // do onthing, just for eliminate warn
    }
})

const onOpenChange = (d:any) => {
    router.push({ path: d.key })
}
</script>
