<template>
    <div>
        <a-modal v-model:visible="visible" :closable="closeable" title="Transaction Detail" width="100%" wrap-class-name="full-modal">
            <template #footer>
                <a-button type="primary" @click="close" v-if="closeable">Close</a-button>
            </template>            
            <h2>Transaction Info</h2> 
            <a-button type="primary" @click="refresh" :loading="loading" class="action-button">Refresh</a-button>      
            <a-popconfirm
                title="Force stop it?"
                ok-text="Yes, stop it"
                ok-type="danger"
                cancel-text="No"
                class="action-button"
                :disabled="transaction?.status==='failed' || transaction?.status==='succeed'"     
                @confirm="handleTransactionStop(<string>transaction?.gid)"                            
            >
                <a-button danger type="default" :disabled="transaction?.status==='failed' || transaction?.status==='succeed'"                                          
                >ForceStop</a-button>
            </a-popconfirm>
            <!-- todo enable condition -->
            <a-popconfirm
                title="Reset next cron time to current time?"
                ok-text="Yes, reset"                
                cancel-text="No"
                class="action-button"                    
                @confirm="handleSetNextCronTimeToNow(<string>transaction?.gid)"            >
                <a-button type="default">Reset next cron time</a-button>
            </a-popconfirm>
            <a-descriptions bordered size="small" :column="{ xxl: 4, xl: 3, lg: 3, md: 3, sm: 2, xs: 1 }">
                <a-descriptions-item label="Status">                          
                    <a-tag :color="transaction?.status === 'succeed' ? 'green' : 'volcano'">{{ transaction?.status }}</a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="Id">{{ transaction?.id }}</a-descriptions-item>
                <a-descriptions-item label="GID">{{ transaction?.gid }}</a-descriptions-item>
                <a-descriptions-item label="TransType">{{ transaction?.trans_type }}</a-descriptions-item> 
                <a-descriptions-item label="Protocol">{{ transaction?.protocol }}</a-descriptions-item> 
                <a-descriptions-item label="CreateTime">{{ transaction?.create_time }}</a-descriptions-item> 
                <a-descriptions-item label="FinishTime">{{ transaction?.finish_time }}</a-descriptions-item> 
                <a-descriptions-item label="UpdateTime">{{ transaction?.update_time }}</a-descriptions-item>                 
                <a-descriptions-item label="NextCronInterval">{{ transaction?.next_cron_interval }}</a-descriptions-item> 
                <a-descriptions-item label="NextCronTime">{{ transaction?.next_cron_time }}</a-descriptions-item> 
                <a-descriptions-item label="RollbackReason">{{ transaction?.rollback_reason }}</a-descriptions-item> 
            </a-descriptions>            
            <h2>Branches</h2>
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
import { useRoute } from 'vue-router';
import { string } from 'vue-types';
import { forceStopTransaction, resetNextCronTime } from '/@/api/api_dtm'
// import VueJsonPretty from 'vue-json-pretty';
// import 'vue-json-pretty/lib/styles.css'
const route = useRoute();

const loading = ref(false)
const dataSource = ref<Branches[]>([])
const transaction = ref<Transaction>()
const visible = ref(false)
const textVal = ref('')
const closeable = ref(true)


let _gid = <string>route.params.gid;
const open = async(gid: string) => {
    _gid = gid;
    loading.value = true;
    const d = await getTransaction<Data>({ gid: gid })
    dataSource.value = d.data.branches
    transaction.value = d.data.transaction
    textVal.value = JSON.stringify(d.data, null, 2)
    visible.value = true
    loading.value = false;
}
if(_gid) {
    open(<string>route.params.gid);
    closeable.value = false;
}

const close = async() => {    
    visible.value = false;
}

const refresh =  async() => {    
    open(_gid);
}

const columns = [
    {
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
    }, {
        title: 'Url',
        dataIndex: 'url',
        key: 'url'
    }
]

const handleTransactionStop = async(gid: string) => {
    await forceStopTransaction(gid);
    refresh();
}


const handleSetNextCronTimeToNow = async(gid: string) => {
    await resetNextCronTime(gid);
    refresh();
}

type Data = {
    branches: {
        gid: string
        branch_id: string
        op: string
        status: string
        create_time: string
        update_time: string
        url: string
    }[]
    transaction: {
        id: number
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
        rollback_reason: string
    }
}

interface Transaction  {
    id: number
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
    rollback_reason: string
}

interface Branches {
    gid: string
    branch_id: string
    op: string
    status: string
    create_time: string
    update_time: string
    url: string
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

  .ant-modal {
    height: 100%;
  }
.action-button {
    margin-right: 10px;
}
</style>
