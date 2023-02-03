<script setup lang="ts">
import Header from './components/header.vue'
</script>

<template>
    <a-layout>
        <a-layout-header class="header">
            <Header />
        </a-layout-header>
        <router-view />
    </a-layout>
</template>
