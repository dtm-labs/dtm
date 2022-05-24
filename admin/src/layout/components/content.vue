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
                        <template v-slot:description>
                            <span v-for="ln of errLines">{{ln}} <br/> </span>
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
const errLines = computed(() => layoutStore.globalError.split("\n"))
const onClose = async (e: MouseEvent) => {
    await sleep(1000)
    layoutStore.setGlobalError("")
}

</script>
