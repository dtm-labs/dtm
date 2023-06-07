<template>
    <div>
        <a-modal v-model:visible="visible" title="Transaction Detail" width="100%" wrap-class-name="full-modal">
            <template #footer>
                <a-button type="primary" @click="close">Close</a-button>
            </template>            
            <a-table :columns="columns" :data-source="dataSource" :pagination="false" :scroll="{ x: true}">
                <!-- eslint-disable-next-line vue/no-unused-vars -->
                <template #bodyCell="{column, record}" />
            </a-table>
            <div class="relative mt-10">
                <a-textarea id="qs" v-model:value="textVal" :auto-size="{ minRows: 10, maxRows: 10 }" />
                <screenfull class="absolute z-50 right-2 top-3" identity="qs" />
            </div>
        </a-modal>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { getTransaction } from '/@/api/api_dtm'
import screenfull from '/@/components/Screenfull/index.vue'
// import VueJsonPretty from 'vue-json-pretty';
// import 'vue-json-pretty/lib/styles.css'

const dataSource = ref<Branches[]>([])
const visible = ref(false)
const textVal = ref('')

const open = async(gid: string) => {
    const d = await getTransaction<Data>({ gid: gid })
    dataSource.value = d.data.branches
    textVal.value = JSON.stringify(d.data, null, 2)
    visible.value = true
}

const close = async() => {    
    visible.value = false;
}

const columns = [
    {
        title: 'GID',
        dataIndex: 'gid',
        key: 'gid'
    }, {
        title: 'BranchID',
        dataIndex: 'branch_id',
        key: 'branch_id'
    }, {
        title: 'Op',
        dataIndex: 'op',
        key: 'op'
    }, {
        title: 'Status',
        dataIndex: 'status',
        key: 'status'
    }, {
        title: 'CreateTime',
        dataIndex: 'create_time',
        key: 'create_time'
    }, {
        title: 'UpdateTime',
        dataIndex: 'update_time',
        key: 'update_time'
    }  
]

type Data = {
    branches: {
        gid: string
        branch_id: string
        op: string
        status: string
        create_time: string
        update_time: string
    }[]
    transactions: {
        ID: number
        create_time: string
        update_time: string
        gid: string
        trans_type: string
        status: string
        protocol: string
        finish_time: string
        options: string
        next_cron_interval: number
        next_cron_time: string
        concurrent: boolean
    }
}

interface Branches {
    gid: string
    branch_id: string
    op: string
    status: string
    create_time: string
    update_time: string
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
