<template>
    <div class="hidden-xs-only px-2">
        <svg-icon v-if="!isFulScreen" class-name="cursor-pointer" icon-class="svg-fullscreen" @click="changeScreenfull(identity)" />
        <svg-icon v-else class-name="cursor-pointer" icon-class="svg-exit-fullscreen" @click="changeScreenfull(identity)" />
    </div>
</template>
<script setup lang='ts'>
import { notification } from 'ant-design-vue'
import { onMounted, onUnmounted, ref } from 'vue'
import screenfull from 'screenfull'

const isFulScreen = ref(false)
const changeScreenfull = (identity: string) => {
    if (!screenfull.isEnabled) {
        notification.open({
            message: '浏览器不支持全屏',
            type: 'warning'
        })
    } else if (identity) {
        const element = document.getElementById(identity)
        screenfull.toggle(element as HTMLElement)
    } else {
        screenfull.toggle()
    }
}
const change = () => {
    if (screenfull.isEnabled) isFulScreen.value = screenfull.isFullscreen
}

defineProps({
    identity: {
        type: String,
        default: null
    }
})

const emits = defineEmits(['screen'])
onMounted(() => screenfull.isEnabled && screenfull.on('change', change) && emits('screen'))
onUnmounted(() => screenfull.isEnabled && screenfull.off('change', change))
</script>
