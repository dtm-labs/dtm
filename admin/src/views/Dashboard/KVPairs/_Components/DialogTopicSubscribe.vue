<template>
    <div>
        <a-modal
            v-model:visible="visible"
            width="60%"
            title="Topic Subscribe"
            :confirm-loading="confirmLoading"
            @ok="handleSubscribe"
        >
            <a-form v-bind="layout" :mode="form">
                <a-form-item label="Topic: ">
                    <a-input v-model:value="form.topic" placeholder="Please input your topic..." />
                </a-form-item>
                <a-form-item label="URL: ">
                    <a-input v-model:value="form.url" placeholder="Please input your url..." />
                </a-form-item>
                <a-form-item label="Remark">
                    <a-textarea v-model:value="form.remark" :rows="6" placeholder="Please input your remark..." />
                </a-form-item>
            </a-form>
        </a-modal>
    </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { reactive, ref } from 'vue'
import { subscribe } from '/@/api/api_dtm'

interface formState {
    topic: string
    url: string
    remark: string
}

const layout = {
    labelCol: { span: 4 },
    wrapperCol: { span: 16 }
}

const form = reactive<formState>({
    topic: '',
    url: '',
    remark: ''
})

const visible = ref(false)
const open = async(topic: string) => {
    form.topic = topic
    visible.value = true
}

const emit = defineEmits(['subscribed'])

const confirmLoading = ref<boolean>(false)
const handleSubscribe = async() => {
    confirmLoading.value = true
    await subscribe<string>(form).then(
        () => {
            visible.value = false
            message.success('Subscribe succeed')
            confirmLoading.value = false
            emit('subscribed')
        }
    )
        .catch(() => {
            message.error('Failed')
            confirmLoading.value = false
            return
        })
}

defineExpose({
    open
})

</script>
