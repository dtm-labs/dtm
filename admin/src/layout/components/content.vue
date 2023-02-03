<template>
    <router-view v-slot="{Component}">
        <transition name="fade-transform" mode="out-in">
            <keep-alive>
                <div>
                    <a-alert
                        v-if="errLines[0]"
                        type="error"
                        closable
                        @close="onClose"
                    >
                        <template #description>
                            <span v-for="(ln, index) of errLines" :key="index">{{ ln }} <br> </span>
                        </template>
                    </a-alert>
                    <component :is="Component" :key="key" />
                </div>
            </keep-alive>
        </transition>
    </router-view>
</template>
<script setup lang='ts'>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { sleep } from '/@/utils/util'
import { useLayoutStore } from '/@/store/modules/layout'

const route = useRoute()

const key = computed(() => route.path)

const layoutStore = useLayoutStore()
const errLines = computed(() => layoutStore.globalError.split('\n'))
const onClose = async() => {
    await sleep(1000)
    layoutStore.setGlobalError('')
}

</script>
