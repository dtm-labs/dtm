import { AxiosResponse } from 'axios'
import request from '/@/utils/request'

export interface IListAllTransactionsReq {
    limit: number
    position?: string
}

export function listAllTransactions<T>(payload: IListAllTransactionsReq): Promise<AxiosResponse<T>> {
    return request({
        url: '/api/dtmsvr/all',
        method: 'get',
        params: payload
    })
}

export function getTransaction<T>(payload: {gid: string}): Promise<AxiosResponse<T>> {
    return request({
        url: '/api/dtmsvr/query',
        method: 'get',
        params: payload
    })
}

export function getDtmVersion(): Promise<AxiosResponse<any>> {
    return request({
        url: '/api/dtmsvr/version',
        method: 'get',
    })
}
