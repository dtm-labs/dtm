<template>
    <div>
        <a-button type="primary" class="mb-2" @click="handleTopicSubscribe('')">Subscribe</a-button>
        <a-table :columns="columns" :data-source="dataSource" :loading="loading" :pagination="false">
            <template #bodyCell="{column, record}">
                <template v-if="column.key === 'subscribers'">
                    <span>{{ JSON.parse(record.v).length }}</span>
                </template>
                <template v-if="column.key === 'action'">
                    <span>
                        <a class="mr-2 font-medium" @click="handleTopicSubscribe(record.k)">Subscribe</a>
                        <a class="mr-2 font-medium" @click="handleTopicDetail(record.k,record.v)">Detail</a>
                        <a class="font-medium text-red-400" @click="handleDeleteTopic(record.k)">Delete</a>
                    </span>
                </template>
            </template>
        </a-table>
        <div v-if="canPrev || canNext" class="flex justify-center mt-2 text-lg pager">
            <a-button type="text" :disabled="!canPrev" @click="handlePrevPage">Previous</a-button>
            <a-button type="text" :disabled="!canNext" @click="handleNextPage">Next</a-button>
        </div>

        <DialogTopicDetail ref="topicDetail" @unsubscribed="handleRefreshData" />
        <DialogTopicSubscribe ref="topicSubscribe" @subscribed="handleRefreshData" />
    </div>
</template>
<script setup lang="ts">
import DialogTopicDetail from './_Components/DialogTopicDetail.vue'
import DialogTopicSubscribe from './_Components/DialogTopicSubscribe.vue'
import { deleteTopic, IListAllKVReq, listKVPairs } from '/@/api/api_dtm'
import { computed, ref } from 'vue-demi'
import { usePagination } from 'vue-request'
import { message, Modal } from 'ant-design-vue'

const columns = [
    {
        title: 'Name',
        dataIndex: 'k',
        key: 'name'
    }, {
        title: 'Subscribers',
        dataIndex: 'v',
        key: 'subscribers'
    }, {
        title: 'Version',
        dataIndex: 'version',
        key: 'version'
    }, {
        title: 'Action',
        key: 'action'
    }
]

const pages = ref([''])
const curPage = ref(1)

const canPrev = computed(() => {
    return curPage.value > 1
})

const canNext = computed(() => {
    return data.value?.data.next_position !== ''
})

type Data = {
    kv: {
        k: string
        v: string
        version: number
    }[]
    next_position: string
}

const queryData = (params: IListAllKVReq) => {
    return listKVPairs<Data>(params)
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const { data, run, current, loading, pageSize } = usePagination(queryData, {
    defaultParams: [
        {
            cat: 'topics',
            limit: 10
        }
    ],
    pagination: {
        pageSizeKey: 'limit'
    }
})

const dataSource = computed(() => data.value?.data.kv || [])

const handlePrevPage = () => {
    curPage.value -= 1
    const params = {
        cat: 'topics',
        limit: pageSize.value,
        position: pages.value[curPage.value] as string
    }
    run(params)
}

const handleNextPage = () => {
    curPage.value += 1
    pages.value[curPage.value] = data.value?.data.next_position as string

    run({
        cat: 'topics',
        position: data.value?.data.next_position,
        limit: pageSize.value
    })
}

const handleRefreshData = () => {
    run({ cat: 'topics', limit: pageSize.value })
}

const handleDeleteTopic = (topic: string) => {
    Modal.confirm({
        title: 'Delete',
        content: 'Do you want delete this topic? ',
        okText: 'Yes',
        okType: 'danger',
        cancelText: 'Cancel',
        onOk: async() => {
            await deleteTopic(topic)
            message.success('Delete topic succeed')
            run({ cat: 'topics', limit: pageSize.value })
        }
    })
}

const topicDetail = ref<null | { open: (topic: string, subscribers: string) => null }>(null)
const handleTopicDetail = (topic: string, subscribers: string) => {
    topicDetail.value?.open(topic, subscribers)
}

const topicSubscribe = ref<null | { open: (topic: string) => null }>(null)
const handleTopicSubscribe = (topic: string) => {
    topicSubscribe.value?.open(topic)
}
</script>

<style lang="postcss" scoped>
::deep .ant-pagination-item {
  display: none;
}

.pager .ant-btn-text {
  font-weight: 500;
}

.pager .ant-btn {
  padding: 6px;
}
</style>
