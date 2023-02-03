<template>
    <div>
        <a-modal v-model:visible="visible" :title="topicName" width="100%" wrap-class-name="full-modal" :footer="null">
            <a-table :columns="columns" :data-source="dataSource" :pagination="false">
                <template #bodyCell="{column, record}">
                    <template v-if="column.key === 'action'">
                        <span>
                            <a class="text-red-400 font-medium" @click="handleUnsubscribe(record.url)">Unsubscribe</a>
                        </span>
                    </template>
                </template>
            </a-table>
            <!-- <div class="mt-10 relative">
          <a-textarea id="qs" v-model:value="textVal" :auto-size="{ minRows: 10, maxRows: 10 }" />
          <screenfull class="absolute right-2 top-3 z-50" identity="qs" />
      </div> -->
        </a-modal>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { unsubscribe } from '/@/api/api_dtm'
import { message, Modal } from 'ant-design-vue'
// import VueJsonPretty from 'vue-json-pretty';
// import 'vue-json-pretty/lib/styles.css'

const dataSource = ref<Subscriber[]>([])
const visible = ref(false)
const topicName = ref<string>('')

const open = async(topic: string, subscribers: string) => {
    dataSource.value = JSON.parse(subscribers)
    topicName.value = topic
    visible.value = true
}

const columns = [
    {
        title: 'URL',
        dataIndex: 'url',
        key: 'url'
    }, {
        title: 'Remark',
        dataIndex: 'remark',
        key: 'remark'
    }, {
        title: 'Action',
        key: 'action'
    }
]

interface Subscriber {
    url: string
    remark: string
}

const handleUnsubscribe = async(url: string) => {
    Modal.confirm({
        title: 'Unsubscribe',
        content: 'Do you want unsubscribe this topic?',
        okText: 'Yes',
        okType: 'danger',
        cancelText: 'Cancel',
        onOk: async() => {
            await unsubscribe({
                topic: topicName.value,
                url: url
            })
            message.success('Unsubscribe topic succeed')
            location.reload()
        }
    })
}

defineExpose({
    open
})

</script>

<style lang="postcss">
.full-modal .ant-modal {
  max-width: 100%;
  top: 0;
  padding-bottom: 0;
  margin: 0;
}

.full-modal .ant-modal-content {
  display: flex;
  flex-direction: column;
  height: calc(100vh);
}

.full-modal .ant-modal-body {
  flex: 1;
}
</style>
