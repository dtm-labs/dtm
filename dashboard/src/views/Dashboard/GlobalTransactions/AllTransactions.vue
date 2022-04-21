<template>
    <div class="w-72">
        <a-input-search v-model:value="value" placeholder="please input gid to search..." enter-button />
    </div>
    <a-table :columns="columns" :data-source="data" class="mt-4">
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
</template>
<script setup lang="ts">
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
const data = [
    {
        key: '1',
        create_time: "2022-04-17T08:56:05+08:00",
        update_time: "2022-04-17T08:56:05+08:00",
        gid: "ZmZPmoqvAhnisrntkELtPb",
        trans_type: "msg",
        status: "succeed",
        protocol: "grpc",
    },
    {
        key: '2',
        create_time: "2022-04-17T08:24:30+08:00",
        update_time: "2022-04-17T08:24:30+08:00",
        gid: "UHTBFL5ukZnrMPwiRQXiNJ",
        trans_type: "msg",
        status: "succeed",
        protocol: "grpc",
    },
    {
        key: '3',
        create_time: "2022-04-17T07:28:30+08:00",
        update_time: "2022-04-17T07:28:30+08:00",
        gid: "2qh5JCfvCaKo3PSxAxrxDn",
        trans_type: "msg",
        status: "succeed",
        protocol: "grpc",
    },
    {
        key: '4',
        create_time: "2022-04-17T07:28:26+08:00",
        update_time: "2022-04-17T07:28:26+08:00",
        gid: "bLvTNErhXJw9YCpQSqZuSH",
        trans_type: "msg",
        status: "succeed",
        protocol: "grpc",
    },
    {
        key: '5',
        create_time: "2022-04-17T07:21:33+08:00",
        update_time: "2022-04-17T07:21:33+08:00",
        gid: "F487jVfVdByDCZMWnBteBE",
        trans_type: "msg",
        status: "succeed",
        protocol: "grpc",
    },
    {
        key: '6',
        create_time: "2022-04-17T07:18:28+08:00",
        update_time: "2022-04-17T07:21:29+08:00",
        gid: "M5yPrCJDYULQg6jS5rqw8C",
        trans_type: "msg",
        status: "succeed",
        protocol: "grpc",
    },
    {
        key: '7',
        create_time: "2022-04-16T11:30:35+08:00",
        update_time: "2022-04-16T11:30:35+08:00",
        gid: "a2JEjm8qMCbiersnaQzTHG",
        trans_type: "msg",
        status: "succeed",
        protocol: "grpc",
    },
    {
        key: '8',
        create_time: "2022-03-19T21:49:13+08:00",
        update_time: "2022-03-19T21:49:13+08:00",
        gid: "4HVsaRkKpWfGsiL6TWso84",
        trans_type: "tcc",
        status: "succeed",
        protocol: "grpc",
    }
]
</script>
