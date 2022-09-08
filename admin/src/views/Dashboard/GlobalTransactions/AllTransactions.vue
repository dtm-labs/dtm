<template>
    <div>
        <a-table :columns="columns" :data-source="dataSource" :loading="loading" :pagination="false">
            <template #bodyCell="{column, record}">
                <template v-if="column.key === 'status'">
                    <span>
                        <a-tag :key="record.status" :color="record.status === 'succeed' ? 'green' : 'volcano'">{{ record.status }}</a-tag>
                    </span>
                </template>
                <template v-else-if="column.key === 'action'">
                    <span>
                        <a class="mr-2 font-medium" @click="handleTransactionDetail(record.gid)">Detail</a>
                        <a-button danger type="link" :disabled="record.status==='failed' || record.status==='succeed'" @click="handleTransactionStop(record.gid)">ForceStop</a-button>
                        <!-- <a class="text-red-400 font-medium"  @click="handleTransactionStop(record.gid)">ForceStop</a> -->
                    </span>
                </template>
            </template>
        </a-table>
        <div class="flex justify-center mt-2 text-lg pager" v-if="canPrev || canNext">
            <a-button type="text" :disabled="!canPrev" @click="handlePrevPage">Previous</a-button>
            <a-button type="text" :disabled="!canNext" @click="handleNextPage">Next</a-button>
        </div>

    <DialogTransactionDetail ref="transactionDetail" />
    </div>
</template>
<script setup lang="ts">
import { IListAllTransactionsReq, listAllTransactions, forceStopTransaction } from '/@/api/api_dtm'
import { ref, computed } from 'vue-demi'
import { usePagination } from 'vue-request'
import DialogTransactionDetail from './_Components/DialogTransactionDetail.vue';
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

const pages = ref([''])
const curPage = ref(1)

const canPrev = computed(() => {
    return curPage.value > 1
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

const queryData = (params: IListAllTransactionsReq) => {
    return listAllTransactions<Data>(params)
}

const { data, run, current, loading, pageSize } = usePagination(queryData, {
    defaultParams: [
        {
            limit: 10,
        }
    ],
    pagination: {
        pageSizeKey: 'limit'
    }
})

const dataSource = computed(() => data.value?.data.transactions || [])

const handlePrevPage = () => {
    curPage.value -= 1;
    const params = {
        limit: pageSize.value,
        position: pages.value[curPage.value] as string
    }
    run(params)
}

const handleNextPage = () => {
    curPage.value += 1;
    pages.value[curPage.value] = data.value?.data.next_position as string

    run({
        position: data.value?.data.next_position,
        limit: pageSize.value,
    })
}

const transactionDetail = ref<null | {open:(gid: string) => null}>(null)
const handleTransactionDetail = (gid: string) => {
    transactionDetail.value?.open(gid)
}

const handleTransactionStop = async (gid: string) => {
    await forceStopTransaction(gid)
    run({
        position: data.value?.data.next_position,
        limit: pageSize.value,
    })
}

</script>

<style lang="postcss" scoped>
::v-deep .ant-pagination-item {
    display: none;
}
.pager .ant-btn-text {
    font-weight: 500;
}
.pager .ant-btn {
    padding: 6px;
}
</style>
