<template>
    <div>
        <a-table :columns="columns" :data-source="dataSource" :loading="loading" :pagination="false">
            <template #bodyCell="{column, record}">
                <template v-if="column.key === 'status'">
                    <span>
                        <a-tag :key="record.status" :color="record.status === 'succeed' ? 'green' : 'volcano'">{{ record.status.toUpperCase() }}</a-tag>
                    </span>
                </template>
                <template v-else-if="column.key === 'action'">
                    <span>
                        <a class="mr-2 font-medium">Detail</a>
                        <a class="text-red-400 font-medium">Stop</a>
                    </span>
                </template>
            </template>
        </a-table>
        <div class="flex justify-center mt-2 text-lg pager" v-if="canPrev || canNext">
            <a-button type="text" :disabled="!canPrev" @click="handlePrevPage">Previous</a-button>
            <a-button type="text" :disabled="!canNext" @click="handleNextPage">Next</a-button>
        </div>
    </div>
</template>
<script setup lang="ts">
import { IListAllTransactions, listAllTransactions } from '/@/api/dtm'
import { ref, onMounted, reactive, computed } from 'vue-demi'
import { usePagination } from 'vue-request'
import { TableProps } from 'ant-design-vue/es/vc-table/Table'
const columns = [
    {
        title: 'GID',
        dataIndex: 'gid',
        key: 'gid'
    }, {
        title: 'TransType',
        dataIndex: 'trans_type',
        key: 'trans_type'
    }, {
        title: 'Status',
        dataIndex: 'status',
        key: 'status'
    }, {
        title: 'Protocol',
        dataIndex: 'protocol',
        key: 'protocol'
    }, {
        title: 'CreateTime',
        dataIndex: 'create_time',
        key: 'create_time'
    }, {
        title: 'Action',
        key: 'action'
    }
]

const pager = ref([""])
const currentState = ref(1)

const canPrev = computed(() => {
    if (currentState.value === 1) {
        return false;
    }

    return true;
})

const canNext = computed(() => {
    return data.value?.data.next_position !== ""
})

type Data = {
    transactions: {
        gid: string
        trans_type: string
        status: string
        protocol: string
        create_time: string
    }[]
    next_position: string
}

const queryData = (params: IListAllTransactions) => {
    return listAllTransactions<Data>(params)
}

const { data, run, current, loading, pageSize } = usePagination(queryData, {
    defaultParams: [
        {
            limit: 100,
        }
    ],
    pagination: {
        pageSizeKey: 'limit'
    }
})

const dataSource = computed(() => data.value?.data.transactions || [])

const handlePrevPage = () => {
    currentState.value -= 1;
    const params = {
        limit: 100
    }
    if (pager.value[currentState.value - 1]) {
        params.position = pager.value[currentState.value - 1]
    }
    run(params)
}

const handleNextPage = () => {
    currentState.value += 1;
    if (currentState.value >= 2) {
        pager.value[currentState.value - 1] = data.value?.data.next_position
    }

    run({
        position: data.value?.data.next_position,
        limit: 5
    })
}
</script>

<style lang="postcss" scoped>
::v-deep .ant-pagination-item {
    display: none;
}
.pager .ant-btn-text {
    font-weight: 500;
}1
.pager .ant-btn {
    padding: 6px;
}
</style>
