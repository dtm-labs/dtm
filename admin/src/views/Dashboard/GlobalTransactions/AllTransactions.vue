<template>
    <div>
        <a-form
            layout="inline"
            :model="{}"
            @finish="searchFinish"
        >
            <a-form-item>
                <a-input v-model:value="gid" placeholder="gid" />
            </a-form-item>
            <a-form-item>
                <a-select
                    ref="select"
                    v-model:value="status"
                    style="width: 200px"
                >
                    <a-select-option value="">-- Status --</a-select-option>
                    <a-select-option value="prepared">prepared</a-select-option>
                    <a-select-option value="submitted">submitted</a-select-option>
                    <a-select-option value="succeed">succeed</a-select-option>
                    <a-select-option value="failed">failed</a-select-option>
                    <a-select-option value="aborting">aborting</a-select-option>
                </a-select>
            </a-form-item>
            <a-form-item>
                <a-select
                    ref="select"
                    v-model:value="transType"
                    style="width: 200px"
                >
                    <a-select-option value="">-- Trans Type --</a-select-option>
                    <a-select-option value="workflow">workflow</a-select-option>
                    <a-select-option value="saga">saga</a-select-option>     
                    <a-select-option value="tcc">tcc</a-select-option>       
                    <a-select-option value="msg">msg</a-select-option>   
                    <a-select-option value="xa">xa</a-select-option>                    
                </a-select>
            </a-form-item>  
            <a-form-item>
                <a-range-picker
                    v-model:value="createTimeRange"
                    format="YYYY-MM-DD HH:mm:ss"
                    :placeholder ="['CreateTime Start', 'CreateTime End']"                    
                />
            </a-form-item>
            <a-form-item>
                <a-button
                    type="primary"
                    html-type="submit"
                >
                    Search
                </a-button>
            </a-form-item>
        </a-form>
    </div>
    <a-divider />
    <div>
        <a-table :columns="columns" :data-source="dataSource" :loading="loading" :pagination="false" :scroll="{ x: true }" size="small" >
            <template #bodyCell="{column, record}">
                <template v-if="column.key === 'status'">
                    <span>
                        <a-tag
                            :key="record.status"
                            :color="record.status === 'succeed' ? 'green' : 'volcano'"
                        >{{ record.status }}</a-tag>
                    </span>
                </template>
                <template v-else-if="column.key === 'action'">
                    <span style="width: 90px; display: block;">
                        <a class="mr-2 font-medium" @click="handleTransactionDetail(record.gid)">Dialog</a>
                        <a class="mr-2 font-medium" target="_blank" :href="'./detail/'+record.gid">Page</a>                        
                    </span>
                </template>
            </template>
        </a-table>
        <div v-if="canPrev || canNext" class="flex justify-center mt-2 text-lg pager">
            <a-button type="text" :disabled="!canPrev" @click="handlePrevPage">Previous</a-button>
            <a-button type="text" :disabled="!canNext" @click="handleNextPage">Next</a-button>
        </div>

        <DialogTransactionDetail ref="transactionDetail" />
    </div>
</template>
<script setup lang="ts">
import { IListAllTransactionsReq, listAllTransactions } from '/@/api/api_dtm'
import { computed, ref } from 'vue-demi'
import { usePagination } from 'vue-request'
import DialogTransactionDetail from './DialogTransactionDetail.vue'

const gid = ref('')
const status = ref('')
const transType = ref('')
const createTimeRange = ref()

const searchFinish = function() {
    curPage.value = 1    
    innerSearch('')
}


const innerSearch = function(position: string) {
    const params = {
        position: position,
        gid: gid.value,
        status: status.value,
        transType: transType.value,
        createTimeStart: createTimeRange.value? createTimeRange.value[0].valueOf(): '',
        createTimeEnd: createTimeRange.value? createTimeRange.value[1].valueOf(): '',
        limit: pageSize.value
    }
    run(params)
}

const columns = [
    {
        title: 'Action',
        key: 'action'
    }, {
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
    },{
        title: 'UpdateTime',
        dataIndex: 'update_time',
        key: 'update_time'
    },{
        title: 'FinishTime',
        dataIndex: 'finish_time',
        key: 'finish_time'
    },{
        title: 'RollbackTime',
        dataIndex: 'rollback_time',
        key: 'rollback_time'
    }, {
        title: 'NextCronInterval',
        dataIndex: 'next_cron_interval',
        key: 'next_cron_interval'
    }, {
        title: 'NextCronTime',
        dataIndex: 'next_cron_time',
        key: 'next_cron_time'
    },
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

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const { data, run, current, loading, pageSize } = usePagination(queryData, {
    defaultParams: [
        {
            limit: 10
        }
    ],
    pagination: {
        pageSizeKey: 'limit'
    }
})

const dataSource = computed(() => data.value?.data.transactions || [])

const handlePrevPage = () => {
    curPage.value -= 1
    let position = pages.value[curPage.value] as string;
    innerSearch(position);
}

const handleNextPage = () => {
    curPage.value += 1
    pages.value[curPage.value] = data.value?.data.next_position as string

    let position = data.value?.data.next_position || '';
    innerSearch(position);  
}

const transactionDetail = ref<null | { open: (gid: string) => null }>(null)
const handleTransactionDetail = (gid: string) => {
    transactionDetail.value?.open(gid)
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
