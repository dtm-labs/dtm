<template>
    <router-view v-slot="{Component}">
        <transition name="fade-transform" mode="out-in">
            <keep-alive>
                <component :is="Component" :key="key" />
            </keep-alive>
        </transition>
    </router-view>
</template>
<script setup lang='ts'>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const key = computed(() => route.path)
</script>
