<template>
    <div>
        <a-layout-header class="header">
            <svg-icon class="logo" style="width: 36px; height: 36px;" icon-class="svg-dtm" />
            <a-menu
            v-model:selectedKeys="activeMenu"
            theme="dark"
            mode="horizontal"
            :style="{ lineHeight: '64px' }"
            @select="onOpenChange"
            >
                <a-menu-item v-for="v in getMenubar.menuList" :key="v.path">{{ v.meta.title }}</a-menu-item>
            </a-menu>
        </a-layout-header>
    </div>
</template>
<script setup lang='ts'>
import { computed, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useLayoutStore } from '/@/store/modules/layout'

const route = useRoute()
const router = useRouter()
const { getMenubar } = useLayoutStore()
const firstRedirectPath = '/dashboard'

const activeMenu = ref([route.meta.activeMenu !== firstRedirectPath ? route.meta.activeMenu : '/'])

const onOpenChange = (d:any) => {
    router.push({ path: d.key })
}
</script>

<style scoped>
.logo {
  float: left;
  margin: 16px 24px 16px 0;
}
</style>
