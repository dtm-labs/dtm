<template>
    <div>
        <a-table :columns="columns" :data-source="dataSource" :pagination="false" :loading="loading" @change="handleTableChange">
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
            limit: 10,
        }
    ],
    pagination: {
        pageSizeKey: 'limit'
    }
})

const dataSource = computed(() => data.value?.data.transactions || [])
const pagination = computed(() => ({
    current: 1,
    pageSize: pageSize.value,
    showSizeChanger: true,

}))

const handleTableChange = (pag: {current:number, pageSize: number}) => {
    if (pag.pageSize !== pageSize.value) {
        run({
            limit: pag.pageSize
        })
    } else {
        run({
            position: data.value?.data.next_position,
            limit: pag.pageSize
        })
    }
}
</script>

<style lang="postcss" scoped>
::v-deep .ant-pagination-item {
    display: none;
}
</style>
